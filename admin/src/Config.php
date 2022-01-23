<?php 
namespace Yjr\A1ertme;
class Config {
	private $cfg, $cfgURI, $redO;
	function __construct($uri) {
		$this->cfg = array();
		$this->cfgURI = $uri;
		$this->redO = new \Redis();
	}
	function connect() {
		if (strlen(trim($this->cfgURI)) == 0) {
			Logger::log("ERROR", "Empty config Redis URI");
			return false;
		};
		$cuParts = explode(":", $this->cfgURI);
		try {
			$_r = sizeof($cuParts) == 2 ? $this->redO->connect($cuParts[0], intval($cuParts[1])) : $this->redO->connect($cuParts[0]);
		} catch (RedisException $e) {
			$_r = false;
		};
		if ($_r == false) {
			Logger::log(ERR, "Failed to connect to config Redis at " . $this->cfgURI);
		};
		return $_r;
	}
	function load() {
		$newCfg = [];
		if ($this->connect()) {
			$newCfg = $this->redO->hGetAll("settings");
			if (count($newCfg) > 0) {
				$this->cfg = $newCfg;
				$this->redO->close();
				return true;
			};
			Logger::log(ERR, "Loaded empty configuration");
			$this->redO->close();
		};
		return false;
	}
	function save() {
		if ($this->connect()) {
			$_r = $this->redO->hMSet("settings", $this->cfg);
			$this->redO->close();
			return $_r;
		};
		return false;
	}
	function get($key) {
		return isset($this->cfg[$key]) ? $this->cfg[$key] : false;
	}
	function set($key, $value) {
		$this->cfg[$key] = $value;
	}
	function dataRedis() {
		# ru - redis URI
		# hps - host+port or socket
		# hp = host+port
		$ru = $this->get("redisURI");
		$hps = false;
		if (!$ru || (preg_match('/^redis:\/\/(.*)$/', $ru, $hps) !== 1 )) {
			return false;
		};
		$ret = new \Redis();
		$hp = false;
		if (preg_match("/([^:]+):(.*)/", $hps[1], $hp) === 1) {
			try {
				Logger::log(DEBUG, $hp[1], intval($hp[2]));
				$_r = $ret->connect($hp[1], intval($hp[2]));
			} catch (Exception $e) {
				$_r = false;
			}
		} else {
			try {
				$_r = $ret->connect($hps[1]);
			} catch (Exception $e) {
				$_r = false;
			}		
		};
		if (!$_r) {
			Logger::log(ERR, "Failed to connect to data Redis at " . $hps[1]);
			return false;
		};
		return $ret;
	}
}
?>