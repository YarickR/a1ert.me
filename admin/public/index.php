<?php
  define("CONFIG_REDIS_URI", getenv("CONFIG_REDIS_URI"));
  require __DIR__ . '/../vendor/autoload.php';
  use Yjr\A1ertme\Logger;
  use Yjr\A1ertme\Config;
  use Yjr\A1ertme\Router;
  use Yjr\A1ertme\Sidebar;
  use Yjr\A1ertme\Main;
  use Yjr\A1ertme\Channel;
  use Yjr\A1ertme\Queue;
  Logger::level(DEBUG);
  Logger::log(INFO, "Starting");
  $cfg = new Config(CONFIG_REDIS_URI);
  if (!$cfg->load()) {
    Logger::log(CRIT, "Failed to load settings from ".CONFIG_REDIS_URI);
    exit();
  };
  Logger::log(DEBUG, var_export($_SERVER, true));
  $routes = 
  [
    "default" => "Yjr\A1ertme\Main::handleDefault",
    "queues" => 
    [  
      "default" => "Yjr\A1ertme\Queue::handleDefault", 
      "unmatched" => "Yjr\A1ertme\Queue::handleUnmatched", 
      "undelivered" => "Yjr\A1ertme\Queue::handleUndelivered"
    ],
    "channels" => 
    [
      "default" => "Yjr\A1ertme\Channel::handleDefault",
      "rules" => "Yjr\A1ertme\Channel::handleRules", 
      "tests" => "Yjr\A1ertme\Channel::handleTests"
    ]
  ];
?>
<html><head><title>a1ert.me management interface</title><link href="/main.css" rel="stylesheet"/></head>
<body id='body'>
<div id='sidebar'><div id='sidebar_header'>Sidebar</div>
<?php echo Sidebar::displaySidebar($routes, Router::getPath($routes, $_SERVER["REQUEST_URI"]));?>
</div>
  <div id='main'><div id='main_header'>Main</div>
<?php 
  $route = Router::getRoute($routes, $_SERVER["REQUEST_URI"]);
  if ($route) {
    call_user_func_array($route, array($cfg));
  }
  ?>
</div>
</body>
</html>

