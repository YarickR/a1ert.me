<?php
  namespace Yjr\A1ertme;
  class Router {
    function __construct() {
    }
    static function getPath($uri) {
      $uriParts = \explode("/", $uri);
      $ret = array();
      $currLevel = $GLOBALS["routes"];
      while (count($uriParts) > 0) {
        $part = trim(array_shift($uriParts));
        if (strlen($part) > 0) {
          if (isset($currLevel[$part])) {
            array_push($ret, $part);
            $currLevel = $currLevel[$part];
          } else {
            break;
          };
        };
      };
      return $ret;
    }
    static function getRoute($uri) {
      $uriParts = \explode("/", $uri);
      $ret = false;
      $currLevel = $GLOBALS["routes"];
      while (count($uriParts) > 0) {
        $part = trim(array_shift($uriParts));
        if (strlen($part) == 0) {
          if (isset($currLevel["default"])) {
            $ret = $currLevel["default"];
          };
          continue;
        };
        if (isset($currLevel[$part])) {
          if (is_array($currLevel[$part])) {
            $currLevel = &$currLevel[$part];
          } else {
            $ret = $currLevel[$part];
            break;
          }
        };
        if (isset($currLevel["default"])) {
          $ret = $currLevel["default"];
        };
      };
      return $ret;
    }
  }
?>
