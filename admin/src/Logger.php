<?php
namespace Yjr\A1ertme;
define("CRIT", 	0);
define("ERR", 	1);
define("WARN", 	2);
define("INFO", 	3);
define("DEBUG", 4);
class Logger {
	private static $currLevel = INFO ;
	private static $pfxs = array (CRIT => "CRITICAL", ERR => "ERROR", WARN => "WARNING", INFO => "INFO", DEBUG => "DEBUG");

	function __construct() {
	}
	public static function level($newLevel) {
		$oldLevel = self::$currLevel;
		self::$currLevel = $newLevel;
		return $oldLevel;		
	} 
	public static function log($level, ...$args) {
		if (($level <= self::$currLevel) && ($level >= CRIT)) {
			$out = "";
            $bt = \debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS, 1);
			foreach ($args as $arg) {
				$out = $out . "\n" .":".var_export($arg, true) ;
			};
			error_log(self::$pfxs[$level].":".$bt[0]["file"].":".$bt[0]["line"].$out);
		}
	}
	public static function showError($out) {
		echo "<script>error(\"$out\");</script>\n";
	}
}	
 ?>
