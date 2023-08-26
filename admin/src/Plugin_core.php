<?php 
namespace Yjr\A1ertme;
class Plugin_core {
  public static function getMenu($cfg) {
    $ret = 
      [
          "default"   => "\Plugin_core::handleChannels",
          "_new"      => "\Plugin_core::handleNew",
          "_edit"     => "\Plugin_core::handleEdit",
          "_save"     => "\Plugin_core::handleSave",
          "_delete"   => "\Plugin_core::handleDelete"

      ];
    return $ret;
  }
  
  public static function handleNew($cfg, $uriPath) {
    ?>
    <form name="new_channel" method="post" action="#">
      <div class="config_new_channel">
        <div class="config_new_channel_label">New channel:</div>
        <div class="config_channel_def">
          <div class="config_channel_def_label">Label:<input type="text" name="new_label"></div>
        </div>
      </div>
      <input type='submit' name='create_new_channel' value='Create channel'>
    </form>
    <?php
  }

  public static function handleEdit($cfg, $uriPath) {
    $chId = -1;
    while (sizeof($uriPath) > 0) {
      $upe = array_shift($uriPath);
      switch ($upe) {
        case "_edit":
          $chId = sizeof($uriPath) > 0 ? intval(array_shift($uriPath)) : -1;
          break;
      };
    };
    if ($chId < 0) { 
      echo "<script>error('Invalid or missing channel id');</script>"; 
      return -1; 
    };
    $chDefs = $cfg->loadPluginConfig(CORE_PLUGIN); /* Channel definitions, indexed by channel id */
    if (!isset($chDefs[$chId])) {
      Logger::log(ERR, "Nonexising channel ".$chId);
      echo "<script>error('Nonexising channel '.$chId);</script>"; 
      return -1;
    };
    $chDef = json_decode($chDefs[$chId], true);
    if (!is_array($chDef)) {
      Logger::log(ERR, "Unable to decode channel ".$chId);
      echo "<script>error('Unable to decode channel '.$chId);</script>"; 
      return -1;
    };
    Logger::log(ERR, var_export($chDef, true));
    $disabledFmt = isset($chDef['format']) ? '' : 'disabled';
    if (!isset($chDef['format'])) {
      $ch0 = json_decode($chDefs[0], true);
      $fmt = $ch0['format'];
    } else {
      $fmt = $chDef['format'];
    };
    if (!is_array($fmt)) {
      $fmt = array(CORE_PLUGIN => $fmt);
    };
    ?>
    Editing channel id <?=$chId;?>
    <div class="config_channel_edit_form">
      <form method="POST" action="#">
        <div class="config_channel_edit_label_div">
          <div class="config_channel_edit_label_label">Label:</div>
          <div class="config_channel_edit_label_value"><input type="text" name="label" size="64" maxlength="128" value="<?=$chDef['label'];?>"></div>
        </div>
        <div class="config_channel_edit_format_div"> 
          <div class="config_channel_edit_format_label">Format:</div>
          <div class="config_channel_edit_format_value">
            <textarea name="format" rows=8 cols=80 <?=$disabledFmt;?>><?=$fmt[CORE_PLUGIN];?></textarea>
            <?php 
              if ($disabledFmt == 'disabled') {
                ?>
                <input type="submit" name="modify" value="modify_format">
                <?php
              }
            ?>
          </div>
        </div> 
        <div class="config_channel_edit_rules_div">
          
        </div>
        <div class="new_entry separator"></div>
        <input type="submit" name="save" value="Save">
      </form>
    </div>
    <?php

  }
  public static function handleSave($cfg, $uriPath) {
    if (isset($_POST['save_changes'])) {
      Logger::log(DEBUG, "Saving modified channels");
    } else
    if (isset($_POST["new_channel"])) {
      $rc = $cfg->connect(); 
      if ($rc != false) {
        $saveRes = $rc->hSetNx("settings_".CORE_PLUGIN, $lastChId == 0 ? 0 : $lastChId + 1, $channelDef);
        if ($saveRes != 1) {
          Logger::showError("Failed to create channel with id ".($lastChId == 0 ? 0 : $lastChId + 1).", please try again");
        };
        $rc->close();
      };
    }
  }

  public static function handleChannels($cfg, $uriPath) {
    $plugCfg = $cfg->loadPluginConfig(CORE_PLUGIN);
    $editLinkBase = "/".implode("/", $uriPath);
    $lastChId = 0;
    if (is_array($plugCfg)) {
      foreach ($plugCfg as $k => $v) {
        if ($k > $lastChId) {
          $lastChId = $k;
        };
      };
    };
    if (isset($_POST["delete_selected"])) {
      $rc = $cfg->connect(); 
      if ($rc != false) {
        Logger::log(DEBUG, "Deleting channels, check references");        
        $rc->close();
      }
    }
    ?><div class="config_channels_header">Channels</div><?php
    if (is_array($plugCfg) || is_object($plugCfg)) {
      foreach ($plugCfg as $k => $v) {
        $chDef = json_decode($v, true);
        ?>
              <div class="config_entry">
                <!--div class="config_channel_def">
                  <div class="config_channel_def_id_div">
                    <div class="config_channel_def_id_label">Id:</div>
                    <div class="config_channel_def_id_value"><?=$k;?></div>
                    <div class="config_edit_channel"><a href="<?=$editLinkBase.'/_edit/'.$k;?>">Edit channel</a></div>
                    <div class="config_delete_channel"><a href="<?=$editLinkBase.'/_delete/'.$k;?>">Delete channel</a></div>
                  </div>
                  <div class="config_channel_def_label_div">
                    <div class="config_channel_def_label_label">Label:</div>
                    <div class="config_channel_def_label_value"><?=$chDef["label"];?></div>
                  </div>
                  <div class="config_channel_def_format_div">
                    <div class="config_channel_def_format_label">Format:</div>
                    <div class="config_channel_def_format_value"><?=$chDef["format"];?></div>
                  </div>
                  <div class="config_channel_def_rules_div">
                    <div class="config_channel_def_rules_label">Rules:</div>
                    <div class="config_channel_def_rules_rule_list">

                  <?php
                    foreach ($chDef["rules"] as $rule) {
                      ?>
                      <div class="config_channel_def_rule">
                        <div class="config_channel_def_rule_src">Src:<?=$rule["src"];?></div>
                        <div class="config_channel_def_rule_cond">Cond:<?=$rule["cond"];?></div>
                      </div>
                      <?php
                    };
                  ?>
                    </div>
                  </div>
                </div-->
                <div class="config_channel_def">
                  <div class="config_channel_def_id_div">
                    <div class="config_channel_def_id_label">Id:</div>
                    <div class="config_channel_def_id_value"><a href="<?=$editLinkBase.'/_edit/'.$k;?>"><?=$k;?></a></div>
                    <div class="config_channel_def_label_label">Label:</div>
                    <div class="config_channel_def_label_value"><?=$chDef["label"];?></div>
                  </div>
                  <div class="config_channel_def_ss_div">
                    Sources: 
                    <?php $srcs = foreach ($chDef["rules"] as $rule) { echo $rule["src"]; }; ?>

                    Sinks:
                  <?php
                    foreach ($chDef["rules"] as $rule) {
                      ?>
                      <div class="config_channel_def_rule">
                        <div class="config_channel_def_rule_src">Src:<?=$rule["src"];?></div>
                        <div class="config_channel_def_rule_cond">Cond:<?=$rule["cond"];?></div>
                      </div>
                      <?php
                    };
                  ?>

                    </div>
                  </div>
                </div>

              </div>
        <?php
      }
    }
    ?>
      <div class="new_entry separator"></div>
      <a href="<?=$editLinkBase.'/_new/';?>">Create channel</a>
    <?php
  }
}
?>