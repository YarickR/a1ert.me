<?php 
namespace Yjr\A1ertme
class Config {
	private $cfg, $cfgURI;

	function __construct($uri) {
		$cfg = [];
		$cfgURI = $uri;
	}
	function load() {
		$newCfg = [];
		$rh = redis_connect($cfgURI);
		if (!$rh) {
			return false;
		};

		$cfg = $newCfg;
		return true;
	}
	function save() {
		$rh = redis_connect($cfgURI);
		if (!$rh) {
			return false;
		};
		
		$cfg = $newCfg;
		return true;

	}
	function get($key) {
		$root = &$cfg;
		$keyParts = split(".", $key);
		while sizeof($keyParts) && $root != false {
			$part = array_shift($keyParts);
			$root = isset($root[$part]) ? &$root[$part] : false;
		};
		return $root;
	}
	function set($key, $value) {
		$root = &$cfg;
		$keyParts = split(".", $key);
		while sizeof($keyParts) && $root != false {
			$part = array_shift($keyParts);
			if (!isset($root[$part])) {
				$root[$part] = sizeof($keyParts) > 1 ? array() : $value;
			} ;
			$root = &$root[$part];
		};
		return true;
	}
}
?>