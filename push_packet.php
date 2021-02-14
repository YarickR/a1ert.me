<?php
  $opts = getopt("h:p:s:f:c:");
  if (!isset($opts['f'])) {
    fwrite(STDERR, "Usage: ".$argv[0]." [-h <host>] [-p <port>] [-s <socket>] [-c <list>] -f <json packet file>\n");
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
    $port = isset($opts['p']) ? int($opts['p']) : 6379;
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
  $contents = json_decode($jStr, true);

  if (!$contents) {
    fwrite(STDERR, "Error parsing json rules file\n");
    exit(4);
  };
  $list = isset($opts['c']) ? $opts['c'] : "alerts";
  $red->rPush($list , $jStr);
?>
