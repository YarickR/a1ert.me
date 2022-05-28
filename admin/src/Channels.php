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
        $newCh = new Channel()
        if ($newCh->init(json_decode($ch, TRUE))->id >= 0) {
          $channels[$newCh->id] = $newCh;
        }
      };
      foreach ($channels as $chId => $ch) {
        foreach ($ch->rules as $rule) {
          if (isset($rule['src'])) {
            array_push($channels[$rule['src']]->sinks, $chId);
          };
        }
      };
      foreach ($channels as $chId => $ch) {
        array_unique($channels[$chId]->sinks);
      };
      return $channels; 
    }
    static function handleDefault($cfg) {
      ?>
        Channels
      <?php
      $chs = Channels::loadChannels($cfg);
      foreach ($chs as $ch) {
        ?>
        <div class='channel_list'>
          <div class='channel_id'><?=$ch->id;?></div>
          <div class='channel_descr'>
            <div class='channel_label'><a href='/channels/modify/?channel_id=<?=$ch->id;?>'><?=$ch->label;?></a></div>
            <div class='channel_sinks'>[ <?php foreach ($ch->sinks as $sink) { printf(" %d ", $sink); };?>]</div>
          </div>
        </div>
        <form name="new_channel" method="POST" action="/channels/modify/">
            <input type="hidden" name="create_channel" value="1">
            <input type="submit" value="Create channel">
        </form>
        <?php
      }
    }
    static function handleTests($cfg) {
      ?>
        Channels
      <?php
    }

  }
?>