<?php
  namespace Yjr\A1ertme;
  class Queues {
    function __construct() {
    }
    public static function getMenu($cfg) {
      $ret = [  
        "default" => "Yjr\A1ertme\Queues::handleDefault", 
        "unmatched" => "Yjr\A1ertme\Queues::handleUnmatched", 
        "undelivered" => "Yjr\A1ertme\Queues::handleUndelivered"
      ];
      return $ret;
    }
    public static function handleDefault($cfg) {
      $dr = $cfg->dataRedis();
      if ($dr) {
        ?>
        <div><?php echo $dr->lLen("alerts");?> pending alerts </div>
        <div><?php echo $dr->lLen("undelivered");?> undelivered alerts </div>
        <div><?php echo $dr->lLen("alerts_backup");?> in backup queue </div>
        <?php
      } else {
        ?>

        <?php
      }
    }
    public static function handleUndelivered($cfg) {
      $dr = $cfg->dataRedis();
      if ($dr) {
        $amt = min($dr->lLen("undelivered"), 10);
        $alerts = $dr->lRange("undelivered", 0, $amt > 0 ? $amt - 1 : 0);
        ?>
        <select name="Undelivered" size="<?=$amt;?>">  
          <?php
            for ($i = 0;$i < count($alerts) ; $i++) {
              Logger::log(DEBUG, $alerts[$i]);
              ?>
              <option><?=$alerts[$i];?></option>
              <?php
            }
          ?>
        </select>  
        <input type="button" label="deliver"/>Try to deliver selected
        <?php
      }
    }
    public static function handleUnmatched($cfg) {

    }

  }  
?>