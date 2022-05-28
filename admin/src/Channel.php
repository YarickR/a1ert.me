<?php
namespace Yjr\A1ertme;
  define("MLOCK", "mlock");
  class Channel {
    public int $id = -1, $version = 0;
    public string $label = "", $group = "";
    public array $rules = [], $sinks = [];
    function __construct() {
    }
    function init($ch) {
      $props = ['id' => true, 'version' => true, 'label' => true, 'group' => false, 'rules' => false];
      foreach ($props as $p => $m) {
        if (!isset($ch[$p]) || ($ch[$p] === false))  {
          if ($m == true) { // Mandatory property 
            error_log(var_export($ch, true));
            error_log("Channel ".var_export($ch, true)." missing mandatory property ".$p); 
            $this->id = -1;
            return $this;
          };
        } else {
          $this->{$p} = $ch[$p];
        };
      };
      return $this;
    };

    function load($id, $rc) {
      $ch = $rc->hGet(CHANNELS, $id);
      if (strlen($ch) > 0) {
        $this->init(json_decode($ch, true));
        return $this->id >= 0;
      };
      return false;
    }
    function save($rc) {
      if ($this->id < 0) {
        error_log(var_export($this, true));
        error_log("Unable to save uninitialised channel");
        return false;
      };
      $props = ['id' => true, 'version' => true, 'label' => true, 'group' => false, 'rules' => false];
      $he = [];
      foreach ($props as $p => $m) {
        $he[$p] = $this->{$p};
      };
      $rc->hSet(CHANNELS, $this->id, json_encode($he));
      return true;
    }
    public static function handleModify($cfg) {
      if (!isset($_GET["channel_id"])) {
        ?>
          <script>err("Missing channel id");</script>
          Channel 
        <?php
          return;
      };
      $ch = new Channel();
      $rc = $cfg->connect();
      if (intval($_GET["channel_id"]) >= 0) {
        $ch->load(intval($_GET["channel_id"]), $rc)
      };

      if (isset($_POST["save"])) {
        if ($rc->hSetNx(CHANNELS, MLOCK, "locked") == true) {
          $ch->save($rc);
          $rc->hDel(CHANNELS, MLOCK);
        }

      }
        $chId = sprintf("channel_%d", $ch->id);
      ?>
        Channel

        <form name="modify_channel" action="#" method="POST">
          <div class='channel_prop'><input type="text" name="<?=$chId;?>_label" value=""></div>
          <div class='channel_prop'><input type="text" name="<?=$chId;?>_group" value=""></div>
        <?php
        foreach ($ch->rules as $rule) {
          $chrId = sprintf("channel_%d_rule_%d", $ch->id, $rule["id"]);
          ?>
          <div class='rule_prop'><input type="text" name="<?=$chrId;?>_src"   value=""></div>
          <div class='rule_prop'><input type="text" name="<?=$chrId;?>_cond"  value=""></div>
          <div class='rule_prop'><input type="text" name="<?=$chrId;?>_link"  value=""></div>
          <input type="submit" value="add_rule"><input type="submit" value="save_channel">
        <?php
        };
        ?>
        </form>

      <?php
    }
  }

?>