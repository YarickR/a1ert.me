<?php
namespace Yjr\A1ertme;
class Logger {
	private $pfxs = [];
	function __construct($uri) {
		define(CRIT, 0);
		define(ERR, 2);
		define(WARN, 2);
		define(INFO, 3);
		define(DEBUG, 4);
		$pfxs = array (CRIT => "CRITICAL", ERR => "ERROR", WARN => "WARNING", INFO => "INFO", DEBUG => "DEBUG");
	}

	public static function log($level, ...$args) {
		$out = "";
		foreach ($arg as $args) {
			$out = $out . "\n". $pfxs[$level].":".var_export($arg, true);
		};
		error_log($out);
	}

}	
 ?>