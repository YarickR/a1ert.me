<?php
  namespace Yjr\A1ertme;
  define("CONFIG_REDIS_URI", getenv("CONFIG_REDIS_URI"));
  define("CONFIG_KEY", getenv("CONFIG_KEY") ? getenv("CONFIG_KEY") : "settings");
  require __DIR__ . '/../vendor/autoload.php';
  use Yjr\A1ertme\Logger;
  use Yjr\A1ertme\Config;
  use Yjr\A1ertme\Router;
  use Yjr\A1ertme\Sidebar;
  use Yjr\A1ertme\Main;
  use Yjr\A1ertme\Channels;
  use Yjr\A1ertme\Queue;
  use Yjr\A1ertme\Plugin_core;
  Logger::level(DEBUG);
  Logger::log(INFO, "Starting");
  $cfg = new Config(CONFIG_REDIS_URI, CONFIG_KEY);
  if (!$cfg->loadMainConfig()) {
    Logger::log(CRIT, "Failed to load settings from ".CONFIG_REDIS_URI);
    exit();
  };
  Logger::log(DEBUG, var_export($_SERVER, true));
  $routes = 
  [
    "default"   => "\Main::handleDefault",
    "config"    => Sidebar::Submenu("\Config::getMenu", $cfg),
//    "queues"    => Sidebar::Submenu("\Queues::getMenu", $cfg),
//    "channels"  => Sidebar::Submenu("\Channels::getMenu", $cfg),
  ];
  $uriPath = Router::getPath($routes, $_SERVER["REQUEST_URI"])
?>
<html><head><title>a1ert.me management interface</title><link href="/main.css?c=<?=rand();?>" rel="stylesheet"/></head>
<script>function error(e) { document.getElementById('errorbar').innerText = e; }</script>
<body id='body'>
<div id='errorbar'></div>
<div id='workarea'>
  <div id='sidebar'><div id='sidebar_header'>Sidebar</div>
    <?php echo Sidebar::displaySidebar($routes, $uriPath);?>
  </div>
  <div class='separator'></div>
  <div id='main'>
    <div id='main_header'>Main</div>
    <div id='main_body'>
<?php 
  $route = Router::getRoute($routes, $_SERVER["REQUEST_URI"]);
  if ($route) {
    call_user_func_array(__NAMESPACE__.$route, array(&$cfg, $uriPath));
  }
  ?>
    </div> <!--main_body-->
  </div> <!--main-->
</div> <!--workarea-->
</body>
</html>

