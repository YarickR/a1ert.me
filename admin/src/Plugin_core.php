<?php 
namespace Yjr\A1ertme;
define("MAX_CHANNEL_LABEL_LEN", 127);   // Arbitrary limitation subject to reevaluation
define("MAX_CHANNEL_FORMAT_LEN", 1023); // Arbitrary limitation subject to reevaluation
define("MAX_CHANNEL_RULE_LEN", 1023);   // Arbitrary limitation subject to reevaluation
class Plugin_core {
  public static function getMenu($cfg) {
    $ret = 
      [
          "default"   => "\Plugin_core::handleChannels",
          "_new"      => "\Plugin_core::handleNewChannel",
          "_edit"     => "\Plugin_core::handleEditChannel",
          "_save"     => "\Plugin_core::handleSaveChannel",
          "_delete"   => "\Plugin_core::handleDeleteChannel",
          "_newrule"  => "\Plugin_core::handleNewRule",
          "_delrule"  => "\Plugin_core::handleDelRule",
      ];
    return $ret;
  }
  
  private static function parseUriPath($uriPath) {
    $ret = [ "prefix" => "/", "chId" => -1 , "rId" => -1];
    while (sizeof($uriPath) > 0) {
      $upe = array_shift($uriPath);
      switch ($upe) {
        case "_edit":
        case "_save":
        case "_delete":
        case "_newrule":
          $ret["chId"] = sizeof($uriPath) > 0 ? intval(array_shift($uriPath)) : -1;
          if ($ret["chId"] == -1) {
            Logger::log(DEBUG, "URI path: {$uriPath}");
          };
          return $ret;
        case "_saverule":
        case "_delrule":
          $ret["chId"] = sizeof($uriPath) > 0 ? intval(array_shift($uriPath)) : -1;
          $ret["rId"] = sizeof($uriPath) > 0 ? intval(array_shift($uriPath)) : -1;
          if (($ret["chId"] == -1) || ($ret["rId"] == -1)) {
            Logger::log(DEBUG, "URI path: {$uriPath}");
          };
          return $ret;
        default:
          $ret["prefix"] .= $ret["prefix"] == "/" ? $upe : "/".$upe;
          break;
      };
    };
    return $ret;
  }

  private static function setupContext($cfg, $uriPath, &$ids, &$chList, &$ch) {
    $ids = self::parseUriPath($uriPath);
    if ($ids["chId"] == -1) {
      Logger::log(ERR, "Malformed URI: ", $uriPath);
      echo "<script>error('Invalid or missing channel id');</script>"; 
      return false;
    };
    $chDefs = $cfg->loadPluginConfig(CORE_PLUGIN); // FIXME: we do not have to pull all the channels to modify only one 
    $chList = new Channels();
    $chList->load($chDefs);
    $ch = $chList->getChannel($ids["chId"]);
    if (!$ch) {
      Logger::log(ERR, "Unknown channel ".$ids["chId"]);
      return false;
    };
    return true;
  } 
  public static function handleNewChannel($cfg, $uriPath) {

    $pathInfo = Plugin_core::parseUriPath($uriPath);
    if (isset($_POST["create_new_channel"]) && isset($_POST["new_label"])) {
      if (strlen(trim($_POST["new_label"])) > 0) {
        $label = substr(trim($_POST["new_label"]), 0, MAX_CHANNEL_LABEL_LEN);
        $plugCfg = $cfg->loadPluginConfig(CORE_PLUGIN);
        $newCh = Channels::createChannel($plugCfg, $label);
        if ($newCh) {
          $cfg->pSet(CORE_PLUGIN, $newCh->id, $plugCfg[$newCh->id]); // FIXME: wildly inefficient and will likely break on large configs
          $cfg->savePluginConfig(CORE_PLUGIN);
          $newUriPath = Router::getPath(sprintf("%s/_edit/%d", $pathInfo["prefix"], $newCh->id));
          return handleEditChannel($cfg, $newUriPath);
        } else {
          echo "<script>error('Unable to create new channel, please try again');</script>"; 
        }
      } else {
        Logger::log(ERR, "Invalid POST data while creating new channel: ", $_POST);
      };
    }
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

  public static function handleEditChannel($cfg, $uriPath) {
    if (!self::setupContext($cfg, $uriPath, $ids, $chList, $ch)) {
      Logger::log(ERR, "Unable to set up channel modification context");
      return -1;
    };
    $pfx = $ids["prefix"];

    $disabledFmt = $ch->format == false  ? 'disabled' : '';
    $fmt = $ch->format == false ? $chList->getChannel(0)->format : $ch->format;
    if (!is_array($fmt)) {
      $fmt = array(CORE_PLUGIN => $fmt);
    };
    ?>
    Editing channel id <?=$ch->id;?>
    <div class="config_channel_edit_form">
      <form name="channel_edit_form" id="channel_edit_form" method="POST" action="<?=sprintf("%s/_save/%d", $pfx, $ch->id);?>">
        <div class="config_channel_edit_label_div">
          <div class="config_channel_edit_label_label">Label:</div>
          <div class="config_channel_edit_label_value"><input type="text" name="label" size="64" maxlength="<?=MAX_CHANNEL_LABEL_LEN;?>" value="<?=$ch->label;?>"></div>
        </div>
        <div class="config_channel_edit_format_div"> 
          <div class="config_channel_edit_format_label">Format:</div>
          <div class="config_channel_edit_format_value">
            <textarea name="format" rows=8 cols=80 <?=$disabledFmt;?>><?=$fmt[CORE_PLUGIN];?></textarea>
            <?php 
              if ($disabledFmt == 'disabled') {
                ?>
                <input type="submit" name="modify_format" value="Modify">
                <?php
              }
            ?>
          </div>
        </div> 
        <div class="config_channel_edit_rules_div">
          <?php
            foreach ($ch->rules as $rId => $rule) {
              $chrulesrc = "rule_{$rId}_src";
              $chrulecond = "rule_{$rId}_cond";
              ?>
                <div class="config_channel_edit_rule_div">
                  <div class="config_channel_edit_rule_src_label">Source</div>
                  <div class="config_channel_edit_rule_src_value">
                    <input type="text" name="<?=$chrulesrc;?>" size="2" maxlength="4" value="<?=$rule["src"];?>">
                  </div>
                  <div class="config_channel_edit_rule_cond_label">Condition</div>
                  <div class="config_channel_edit_rule_cond_value">
                    <input type="text" name="<?=$chrulecond;?>" size="60" maxlength="1023" value="<?=$rule["cond"];?>">
                  </div>
                  <div class="config_channel_edit_rule_delete">
                      <a  href="<?=sprintf("%s/_delrule/%d/%d", $pfx, $ch->id, $rId);?>" 
                          onclick="cl('Do you want to delete rule <?=$rId;?> ?', this);">Delete rule</a>
                  </div>
                </div>
              <?php
            };
          ?>
        </div>
        </form>
        <div class="new_entry separator"></div>
        <form name="new_rule_form" id="new_rule_form" action="<?=sprintf("%s/_newrule/%d", $pfx, $ch->id);?>" method="POST">
        <div class="config_channel_new_rule">
          <div class="config_channel_new_rule_src_label">Source channel:</div><input type="text" name="new_rule_src" size="2" maxlength="4">
          <div class="config_channel_new_rule_src_label">Condition:</div><input type="text" name="new_rule_cond" size="60" maxlength="1023">
          <input type="button" name="add_rule_button" value="Add rule" onclick='document.forms["new_rule_form"].submit();'>
        </form>
        </div>
        <input type="button" name="save_channel_button" value="Save changes" onclick='document.forms["channel_edit_form"].submit();'>
    </div>
    <?php

  }
  public static function handleSaveChannel($cfg, $uriPath) {
    if (!self::setupContext($cfg, $uriPath, $ids, $chList, $ch)) {
      Logger::log(ERR, "Unable to set up channel modification context");
      return -1;
    };
    $pfx = $ids["prefix"];
    if (isset($_POST["save_channel"])) {
      Logger::log(DEBUG, "Saving modified channel");
      if (strlen(trim($_POST["label"]) > 0)) {
        $ch->label = trim($_POST["label"]);
      } else {
        Logger::log(ERR, "Invalid label in POST data:", $_POST);
        echo "<script>error('Invalid label');</script>";
      };
      foreach ($ch->rules as $rId => $rule) {
        $rSrcName = sprintf("rule_%d_src");
        $rCondName = sprintf("rule_%d_cond");
        $newRule = array(
            "src"   => intval($_POST["rule_{$rId}_src"]), 
            "cond"  => trim($_POST["rule_{$rId}_cond"])); 
        if ($chList->verifyRule($newRule)) {
          $ch->rule[$rId] = $newRule;
        } else {
          Logger::log(ERR, "Invalid rule form data: ", $_POST);
          echo "<script>error('Invalid rule {$rId}');</script>";
        };
      };
      $chList->setChannel($ch);
      $cfg->pSet(CORE_PLUGIN, $ch->id, $ch->save()); 
      $cfg->savePluginConfig(CORE_PLUGIN); // FIXME: save only modified values  
    } else 
    if (isset($_POST["modify_format"])) {
      Logger::log(DEBUG, "COWing format");
      if (strlen(trim($_POST["format"])) > 0) {
        $ch->format = trim($_POST["format"]);
      } else {
        $ch->format = false;
      };
      $chList->setChannel($ch);
      $cfg->pSet(CORE_PLUGIN, $ch->id, $ch->save());
      $cfg->savePluginConfig(CORE_PLUGIN); // FIXME: save only modified values  
    };
    $newUriPath = Router::getPath(sprintf("%s/_edit/%d", $pfx, $ch->id));
    return handleEditChannel($cfg, $newUriPath);
  }

  public static function handleChannels($cfg, $uriPath) {
    $ids = Plugin_core::parseUriPath($uriPath);
    $pfx = $ids["prefix"];
    $chDefs = $cfg->loadPluginConfig(CORE_PLUGIN); // FIXME 
    $lastChId = 0;
    $chList = new Channels();
    $chList->load($chDefs);

    ?><div class="config_channels_header">Channels</div><?php
    $chIds = $chList->ids();
    foreach ($chIds as $chId) {
      $chDef = &$chList->getChannel($chId);
      ?>
        <div class="config_entry">
          <div class="config_chlist">
            <div class="config_chlist_edit">
                <a href="<?=sprintf('%s/_edit/%d', $pfx, $chId);?>">Edit</a>
            </div>
            <div class="config_chlist_id">
              <div class="config_chlist_id_label">Id:</div>
              <div class="config_chlist_id_value"><?=$chId;?></div>
            </div>
            <div class="config_chlist_label">
              <div class="config_chlist_label_label">Label:</div>
              <div class="config_chlist_label_value"><?=$chDef->label;?></div>
            </div>
            <div class="config_chlist_sources">
              <div class="config_chlist_sources_label">Sources:</div>
              <div class="config_chlist_sources_value">
                <?php 
                  foreach ($chDef->sources as $srcId) {
                    $srcCh = $chList->getChannel($srcId);
                    if ($srcCh) {
                      echo "<a href='#'>{$srcId} - {$srcCh->label}</a>";
                    };
                  };
                ?>
              </div>
            </div>
            <div class="config_chlist_sinks">
              <div class="config_chlist_sinks_label">Sinks:</div>
              <div class="config_chlist_sinks_value">
                <?php 
                  foreach ($chDef->sinks as $sinkId) {
                    $sinkCh = &$chList->getChannel($sinkId);
                    if ($sinkCh) {
                      echo "<a href='#'>{$sinkId} - {$sinkCh->label}</a>";
                    };
                  };
                ?>
              </div>
            </div>
            <div class="config_chlist_delete">
                <a href="<?=sprintf('%s/_delete/%d', $pfx, $chId);?>" onclick="cl('Do you want to delete channel <?=$chId;?> ?', this);">Delete</a>
            </div>
          </div>
        </div>
      <?php
    }
    ?>
      <div class="new_entry separator"></div>
      <a href="<?=sprintf('%s/_new/', $pfx);?>">Create channel</a>
    <?php
  }

  public static function handleNewRule($cfg, $uriPath) {
    if (!self::setupContext($cfg, $uriPath, $ids, $chList, $ch)) { // FIXME: 
      Logger::log(ERR, "Unable to set up channel modification context");
      return -1;
    };
    $pfx = $ids["prefix"];
    if (isset($_POST["new_rule_src"]) && isset($_POST["new_rule_cond"])) {
      $newRule = array("src" => intval($_POST["new_rule_src"]), "cond" => trim($_POST["new_rule_cond"]));
      if ($chList->validateRule($newRule) && $ch->createRule($newRule)) {
        $chList->setChannel($ch);
        $cfg->pSet(CORE_PLUGIN, $ch->id, $ch->save());
        $cfg->savePluginConfig(CORE_PLUGIN);
      } else {
        Logger::log(ERR, "Invalid rule data : ", $_POST);
      }
    };
    $newUriPath = Router::getPath(sprintf("%s/_edit/%d", $pfx, $ch->id));
    return handleEditChannel($cfg, $newUriPath);
  }

  public static function handleDelRule($cfg, $uriPath) {
    if (!self::setupContext($cfg, $uriPath, $ids, $chList, $ch)) { // FIXME: 
      Logger::log(ERR, "Unable to set up channel modification context");
      return -1;
    };
    if ($chList->canDeleteRule($ch->id, $ids["rId"]) && $ch->deleteRule($ids["rId"])) {
      $chList->setChannel($ch);
      $cfg->pSet(CORE_PLUGIN, $ch->id, $ch->save());
      $cfg->savePluginConfig(CORE_PLUGIN);
    } else {
      Logger::log(ERR, "Unable to delete rule".$ids["rId"]." in channel ".$ch->id);
    };
    $newUriPath = Router::getPath(sprintf("%s/_edit/%d", $pfx, $ch->id));
    return handleEditChannel($cfg, $newUriPath);
  }
}
?>