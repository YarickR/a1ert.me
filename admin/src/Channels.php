<?php
  namespace Yjr\A1ertme;
  define('CHANNELS', "channel_defs");
  class Channels {
    function __construct() {
    }
    static function loadChannels($cfg) {
      $channels = array();
      $rc = $cfg->connect();
      if (!$rc) {
        return $channels;
      };
      
      $chList = $rc->hGetAll(CHANNELS);
      foreach ($chList as $ch) {
        $newCh = new Channel();
        if ($newCh->init(json_decode($ch, TRUE))->id >= 0) {
          $channels[$newCh->id] = $newCh;
        }
      };
      foreach ($channels as $chId => $ch) {
        $chSrcs = array();
        foreach ($ch->rules as $rule) {
          if (is_array($rule['src'])) {
            $chSrcs = array_merge($chSrcs, $rule['src']); 
          } else {
            array_push($chSrcs, intval($rule['src']));
          };
        };
        $chSrcs = array_unique($chSrcs, SORT_NUMERIC);
        asort($chSrcs, SORT_NUMERIC);
        foreach ($chSrcs as $srcId) {
          array_push($channels[intval($srcId)]->sinks, intval($chId));
        };
        $channels[$chId]->srcs = $chSrcs;
      };
      foreach ($channels as $chId => $ch) {
        array_unique($channels[$chId]->sinks);
        asort($channels[$chId]->sinks, SORT_NUMERIC);
      };
      ksort($channels, SORT_NUMERIC);
      return $channels; 
    }
    public static function getMenu($cfg) {
      $ret =     
        [
          "default" => "Channels::handleDefault",
          "_modify" => "Channel::handleModify", 
          "test" => "Channels::handleTest"
        ];
        return $ret;
    }
    static function handleDefault($cfg) {
      ?>
        Channels
      <?php
      $chs = Channels::loadChannels($cfg);
      ?>
      <div class='channel_list'>
      <?php
      foreach ($chs as $ch) {
        ?>
        <div class='channel'>
          <div class='channel_id'><?=$ch->id;?></div>
          <div class='channel_descr'>
            <div class='channel_sinks'>[ <?php foreach ($ch->srcs as $src) { printf(" %d ", $src); };?>]</div>
            <div class='channel_label'> -> <a href='/channels/_modify/?channel_id=<?=$ch->id;?>'><?=$ch->label;?></a> -> </div>
            <div class='channel_sinks'>[ <?php foreach ($ch->sinks as $sink) { printf(" %d ", $sink); };?>]</div>
          </div>
        </div>
        <?php
      }
      ?>
        <div class='channel'>
          <div class='channel_id'>
            <form name="new_channel" method="POST" action="/channels/_modify/">
              <input type="hidden" name="create_channel" value="1">
              <input type="submit" value="Create channel">
            </form>
          </div>
          <div class='channel_descr'>
          </div>
        </div>
      </div>
      <?php
    }
    static function handleTests($cfg) {
      ?>
        Channels
      <?php
    }

  }
?>