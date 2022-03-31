<?php
  $opts = getopt("h:p:s:f:");
  if (!isset($opts['f'])) {
    fwrite(STDERR, "Usage: ".$argv[0]." [-h <host>] [-p <port>] [-s <socket>] -f <json rules file>\n");
    exit(1);
  };
  $rulesFile = fopen($opts['f'], "r");
  if ($rulesFile == FALSE) {
    fwrite(STDERR, "Cannot open ".$opts['f']." for reading\n");
    exit(2);
  };
  fclose($rulesFile); // we will do file_get_contents later , but we need to bail early if the file is not availabe
  $red = new Redis();

  if (isset($opts['h'])) {
    $port = isset($opts['p']) ? intval($opts['p']) : 6379;
    if ($red->connect($opts['h'], $port) == FALSE) {
      fwrite(STDERR, "Error connecting to Redis at ".$opts['h'].":".$port."\n");
      exit(3);
    }
  } else 
  if (isset($opts['s'])) {
    if ($red->connecT($opts['s']) == FALSE) {
      fwrite(STDERR, "Error connecting to Redis at ".$opts['s']."\n");
      exit(3);
    }
  };
  $jStr = file_get_contents($opts['f']);
  $rules = json_decode($jStr, true);

  if (!$rules) {
    fwrite(STDERR, "Error parsing json rules file\n");
    exit(4);
  };
  $currChVer = $red->hGet("settings", "channel_version");
  $currChVer = $currChVer == FALSE ? 1 : intval($currChVer);

  function idcmp($a, $b) {
    if ($a['id'] == $b['id']) return 0;
    return $a['id'] < $b['id'] ? -1 : 1;
  };

  usort($rules, "idcmp");
  $lastChId = 0;
  $rm = $red->multi();
  $rm->del("channel_defs", 1, 0);
  foreach($rules as $rule) {
    if ($rule['id'] > $lastChId) {
      $lastChId = $rule['id'];
    };
    $rm->hSet("channel_defs", $rule["id"], json_encode($rule, JSON_NUMERIC_CHECK|JSON_UNESCAPED_SLASHES|JSON_UNESCAPED_UNICODE));
    fwrite(STDOUT, json_encode($rule, JSON_NUMERIC_CHECK|JSON_UNESCAPED_SLASHES|JSON_UNESCAPED_UNICODE)."\n");
  };
  $rm->hSet("settings", "channel_version", $currChVer + 1);
  $rm->hSet("settings", "last_channel_id", $lastChId);
  $rm->exec();
  $red->bgSave();
?>
