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
		}
		$cuParts = explode(":", $this->cfgURI);
		try {
			if (sizeof($cuParts) == 2) {
				// assuming host:port
				$this->redO->connect($cuParts[0], intval($cuParts[1]));
			} else {
				// this could be unix socket
				$this->redO->connect($cuParts[0]);
			}
		} catch (Exception $e) {
			Logger::log(ERR, "Exception ".var_export($e, true)." connecting to ".$this->cfgURI);
			return false;
		}
		return true;
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
	 		return false;
		}
	}
	function save() {
		if ($this->connect()) {
			$ret = $this->redO->hMSet("settings", $this->cfg);
			$this->redO->close();
			return $ret;
		};
		return false;
	}
	function get($key) {
		return isset($this->cfg[$key]) ? $this->cfg[$key] : false;
	}
	function set($key, $value) {
		$this->cfg[$key] = $value;
	}
	function getDataRedis() {
		$redisURI = $this->get("redisURI");
		if ($redisURI) {
			$hps = preg_split("redis://(.*)", $redisURI);
			if ($hps[1]) {
				
			}
			$drc = new \Redis();
			try {
				if (sizeof($cuParts) == 2) {
					// assuming host:port
					$this->redO->connect($cuParts[0], intval($cuParts[1]));
				} else {
					// this could be unix socket
					$this->redO->connect($cuParts[0]);
				}
			} catch (Exception $e) {
				Logger::log(ERR, "Exception ".var_export($e, true)." connecting to ".$this->cfgURI);
				return false;
			}
			return true;
		}
	}
}
?>