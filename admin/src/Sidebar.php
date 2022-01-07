<?php
  namespace Yjr\A1ertme;
  class Sidebar {
    function __construct() {

    }
    static function displaySidebarEntry($pre, $path, $routes, $key, $level, $fullPath, $onPath) {
      if ($key == "default") {
        return $pre;
      };
      $path = sprintf("%s/%s", $path, $key); 
      $onPath = $onPath & ((count($fullPath) > $level) && ($fullPath[$level] == $key));
      $pathMark = $onPath ? "*" : "";
      $pre = $pre . sprintf("<div><div class='sidebar_level_%d'><a href='%s'>%s%s</a></div></div>\n", $level, $path, $pathMark, $key);
      if (is_array($routes[$key])) {
        foreach ($routes[$key] as $chK => $chR) {
          $pre = self::displaySidebarEntry($pre, $path, $routes[$key], $chK, $level + 1, $fullPath, $onPath);
        };
      };
      return $pre;
    }
    static function displaySidebar($routes, $currPath) {
      $ret = []; 
      foreach ($routes as $rK => $r) {
        array_push($ret, self::displaySidebarEntry("", "", $routes, $rK, 0, $currPath, true));
      };
      return implode("\n", $ret);
    }
  }
?>