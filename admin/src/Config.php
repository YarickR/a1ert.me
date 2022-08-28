<?php 
namespace Yjr\A1ertme;
class Config {
	private $cfg, $cfgURI, $cfgKey, $redO;
	function __construct($uri, $key) {
		$this->cfg = array();
		$this->cfgURI = $uri;
		$this->cfgKey = $key;
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
		} catch (\RedisException $e) {
			$_r = false;
		};
		if ($_r == false) {
			Logger::log(ERR, "Failed to connect to config Redis at " . $this->cfgURI);
		};
		return $_r ? $this->redO : false;
	}
	function load() {
		$newCfg = [];
		$rc = $this->connect(); 
		if ($rc != false) {
			$newCfg = $rc->hGetAll($this->cfgKey);
			if (count($newCfg) > 0) {
				$this->cfg = $newCfg;
				$rc->close();
				return true;
			};
			Logger::log(ERR, "Loaded empty configuration");
			$rc->close();
		};
		return false;
	}
	function save() {
		$rc = $this->connect();
		if ($rc != false) {
			$_r = $rc->hMSet($this->cfgKey, $this->cfg);
			$rc->close();
			return $_r;
		};
		return false;
	}
	function get($key) {
		return isset($this->cfg[$key]) ? $this->cfg[$key] : false;
	}
	function keys() {
		return array_keys($this->cfg);
	}
	function set($key, $value) {
		$this->cfg[$key] = $value;
	}

	function getPlugins() {
		$ps = $this->get("plugins");
		return $ps && (strlen($ps) > 0) ? preg_split("/[\ ,:\/]+/", $ps, -1, PREG_SPLIT_NO_EMPTY) : array();
	}
	function loadPluginConfig($plugin) {
		$newCfg = [];
		$rc = $this->connect(); 
		if ($rc != false) {
			$newCfg = $rc->hGetAll($this->cfgKey."_".$plugin);
			if (count($newCfg) > 0) {
				$rc->close();
				return $newCfg;
			};
			Logger::log(ERR, "Loaded empty configuration for ".$plugin." plugin");
			$rc->close();
		};
		return false;
	}
	function savePluginConfig($plugin, $pluginCfg) {
		$newCfg = [];
		$rc = $this->connect(); 
		if ($rc != false) {
			$s = $rc->hMSet($this->cfgKey."_".$plugin, $pluginCfg);
			$rc->close();
		};
		return false;
	}

	function dataRedis() {
		# ru - redis URI
		# hps - host+port or socket
		# hp = host+port
		$ru = $this->get("data_redis");
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
	public static function handleSave($cfg) {
		$cfg->save();
		Config::handleDefault($cfg);
	}

	public static function getMenu($cfg) {
		$ret = 
		[
      		"default" => "Config::handleDefault",
      		"_save"   => "Config::handleSave"
    	];
    	$plugins = $cfg->getPlugins();
		foreach ($plugins as $p) {
			$ret[$p] = "Config::handlePluginConfig";
    	};
    	return $ret;
	}
	public static function handlePluginConfig($cfg, $uriPath) {
		$plugin = $uriPath[sizeof($uriPath) - 1]; // last entry
		$plugCfg = $cfg->loadPluginConfig($plugin);
		if (isset($_POST['save'])) {
			foreach ($_POST as $k => $v) {
				if (preg_match('/^value_(.+)$/', $k, $m) === 1) {
					$plugCfg[$m[1]] = $v;
				};
			};
			$cfg->savePluginConfig($plugin, $plugCfg);
		};
		?>Config
		<form name="config" method="post" action="#"><?php
		foreach ($plugCfg as $k => $v) {
			?>
        		<div class="config_entry"><div class="config_key"><?=$k;?></div><div class="config_value"><input type="text" name="value_<?=$k?>" value="<?=$v;?>"></div></div>
			<?php
		}
		?>
		<input type="submit" name="save" value="Save">
		</form>

		<?php
	}
	public static function handleDefault($cfg) {
		if (isset($_POST['save'])) {
			foreach ($_POST as $k => $v) {
				if (preg_match('/^value_(.+)$/', $k, $m) === 1) {
					$cfg->set($m[1], $v);
				};
			};
			$cfg->save();
		};
		?>Config
		<form name="config" method="post" action="#"><?php
		foreach ($cfg->keys() as $k) {
			?>
        		<div class="config_entry"><div class="config_key"><?=$k;?></div><div class="config_value"><input type="text" name="value_<?=$k?>" value="<?=$cfg->get($k);?>"></div></div>
			<?php
		}
		?>
		<input type="submit" name="save" value="Save">
		</form>
		<?php
    }

}
?>