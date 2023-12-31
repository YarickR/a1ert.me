<?php 
namespace Yjr\A1ertme;
define("CORE_PLUGIN", "core");
class Config {
	private $cfg, $cfgURI, $cfgKey, $redO;
	function __construct($uri, $key) {
		$this->cfg = array();
		$this->cfgURI = $uri;
		$this->cfgKey = $key;
        $this->plugins = array();
		$this->redO = new \Redis();
	}
	function connect() {
		if (strlen(trim($this->cfgURI)) == 0) {
			Logger::log(ERR, "Empty config Redis URI");
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
	function loadMainConfig() {
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
	function saveMainConfig() {
		$rc = $this->connect();
		if ($rc != false) {
			$_r = $rc->hMSet($this->cfgKey, $this->cfg);
			$rc->close();
			return $_r;
		};
		return false;
	}
	function mGet($key) {
		return isset($this->cfg[$key]) ? $this->cfg[$key] : false;
	}
	function mKeys() {
		return array_keys($this->cfg);
	}
	function mSet($key, $value) {
		$this->cfg[$key] = $value;
	}

	function getPlugins() {
		$ps = $this->mGet("plugins");
        $plugList = array();
        if ( $ps && (strlen($ps) > 0)) {
            $plugList = preg_split("/[\ ,:\/]+/", $ps, -1, PREG_SPLIT_NO_EMPTY);
        };
        array_unshift($plugList, CORE_PLUGIN);
        $plugList = array_unique($plugList);
        foreach ($plugList as $plugName) {
            if (!isset($this->plugins[$plugName])) {
                $this->plugins[$plugName] = array();
            };
        };
        return $plugList;
	}

	function loadPluginConfig($plugin) {
		$newCfg = [];
		$rc = $this->connect(); 
		if ($rc != false) {
			$newCfg = $rc->hGetAll($this->cfgKey."_".$plugin);
			if (count($newCfg) > 0) {
				$rc->close();
                $this->plugins[$plugin] = $newCfg;
 				return $newCfg;
			};
			Logger::log(ERR, "Loaded empty configuration for ".$plugin." plugin");
			$rc->close();
		};
		return false;
	}
	function savePluginConfig($plugin) { // One bix FIXME, need to support robust concurrent modification of key value pairs
		if (!isset($this->plugins[$plugin])) {
			Logger::log(ERR, "Missing {$plugin} config in cached data");
			return false;
		};
		$rc = $this->connect();
		if ($rc != false) {
			$s = $rc->hMSet($this->cfgKey."_".$plugin, $this->plugins[$plugin]);
			return $s;
		};
		return false;
	}

    function pGet($plugin, $key) {
        return isset($this->plugins[$plugin]) && isset($this->plugins[$plugin][$key]) ? $this->plugins[$plugin][$key] : false;
    }
    function pSet($plugin, $key, $value) {
        if (isset($this->plugins[$plugin])) {
            $this->plugins[$plugin][$key] = $value;
            return true;
        };
        return false;
    }
    function pKeys($plugin) {
		return isset($this->plugins[$plugin]) ? array_keys($this->plugins[$plugin]) : false;
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
      		"default" => "\Config::handleDefault",
      		"_save"   => "\Config::handleSave"
    	];
        $plugins = $cfg->getPlugins();
		foreach ($plugins as $p) {
			if (method_exists(__NAMESPACE__."\Plugin_".$p, "getMenu")) {
				$ret[$p] = call_user_func(__NAMESPACE__."\Plugin_".$p."::getMenu", $cfg);
			} else {
				$ret[$p] = "\Config::handlePluginConfig";
        	}
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
		} else
		if (isset($_POST["new"])) {
			$rc = $cfg->connect(); 
			if ($rc != false) {
		  	$nf = trim($_POST["new_key"]);
				if ((strlen($nf) > 0) && !isset($plugCfg[$nf]) && (strlen(trim($_POST["new_value"])) > 0)) {
					$plugCfg[$nf] = trim($_POST["new_value"]);
					$s = $rc->hSet($cfg->cfgKey."_".$plugin, $nf, $plugCfg[$nf]);
					$rc->close();
				};
			};
		} else 
		if (isset($_POST["delete"])) {
			$rc = $cfg->connect(); 
			if ($rc != false) {
				foreach ($_POST as $pk => $pv) {
					if (preg_match('/delete_(.*)$/', $pk, $pkm) === 1) {
						unset($plugCfg[$pkm[1]]);
						$rc->hDel($cfg->cfgKey."_".$plugin, $pkm[1]);
					};
				};
				$rc->close();
			}
		}
		?>Config
		<form name="config" method="post" action="#"><?php
		if (is_array($plugCfg)  || is_object($plugCfg)) {
			foreach ($plugCfg as $k => $v) {
				?>
	        		<div class="config_entry">
	        			<input type="checkbox" class="config_select" name="delete_<?=$k?>" value="0">
	        			<div class="config_key"><?=$k;?></div>
	        			<div class="config_value"><input type="text" name="value_<?=$k?>" value="<?=$v;?>"></div>
	        		</div>
				<?php
			};
		};
		?>
    	<div class="config_save_delete">
    		<input type="submit" name="delete" class="config_save_delete_delete" value="Delete Selected">
			<input type="submit" name="save" class="config_save_delete_save" value="Save Changes">
		</div>
		<div class="new_entry separator"></div>
		<div class="config_add_entry">
			<div class="config_new_key_label">New key:</div><input type="text" class="config_new_key_input" name="new_key">
		    <div class="config_new_value_label">New value:</div><input type="text" class="config_new_value_input" name="new_value" value="">
		    <input type="submit" class="config_new_value_submit" name="new" value="Add">
		</div>
		</form>

		<?php
	}
	public static function handleDefault($cfg) {
		if (isset($_POST['save'])) {
			foreach ($_POST as $k => $v) {
				if (preg_match('/^value_(.+)$/', $k, $m) === 1) {
					$cfg->mSet($m[1], $v);
				};
			};
			$cfg->saveMainConfig();
		};
		?>Config
		<form name="config" method="post" action="#"><?php
		foreach ($cfg->mKeys() as $k) {
			?>
        		<div class="config_entry"><div class="config_key"><?=$k;?></div><div class="config_value"><input type="text" name="value_<?=$k?>" value="<?=$cfg->mGet($k);?>"></div></div>
			<?php
		}
		?>
		<input type="submit" name="save" value="Save">
		</form>
		<?php
    }
}
?>
