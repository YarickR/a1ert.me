<?php
  define(REDIS_URI, getenv("REDIS_URI"));
  require __DIR__ . '/../vendor/autoload.php';
  use Yjr\A1ertme\Logger;
  use Yjr\A1ertme\Config;
  $cfg = new Config();
  if !$cfg::load(REDIS_URI) {
    exit();
  }

?>
<html><head><title>a1ert.me management interface</title></head>
<body>
</body>
</html>

