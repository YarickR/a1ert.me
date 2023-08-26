<?php
  namespace Yjr\A1ertme;
  class Channels { // FIXME rework to not use the whole plugCfg
    private $channels;
    private $noChannel = false;
    function __construct() {
      $this->channels = array();
    }
    function load(&$plugCfg) {
      $this->channels = array();
      foreach ($plugCfg as $chDefId => $chDef) {
        $newCh = new Channel();
        if ($newCh->load($chDef)) {
          if (intval($newCh->id) != intval($chDefId)) { // Should not happen
            Logger::log(ERR, "Channel Ids do not match: {$newCh->id} vs {$chDefId}");
          } else {
            $this->channels[intval($newCh->id)] = $newCh;
          };
        };
      };
      foreach ($this->channels as $chId => $ch) {
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
          array_push($this->channels[$srcId]->sinks, intval($chId));
        };
        $this->channels[$chId]->sources = $chSrcs;
      };
      foreach ($this->channels as $chId => $ch) {
        array_unique($this->channels[$chId]->sinks);
        asort($this->channels[$chId]->sinks, SORT_NUMERIC);
      };
      ksort($this->channels, SORT_NUMERIC);
      return sizeof($this->channels); 
    }
    function ids() {
      return array_keys($this->channels);
    }
    function getChannel($id) {
      $ret = isset($this->channels[intval($id)]) ? $this->channels[intval($id)] : $this->noChannel;
      return $ret;
    }
    function setChannel($channel) {
      if (is_object($channel) && isset($this->channels[$channel->id])) {
        $this->channels[$channel->id] = $channel;
        return true;
      };
      return false;
    }
    function save(&$plugCfg) {
      foreach ($this->channels as $chId => $channel) {
        $plugCfg[$chId] = $channel->save(); // FIXME
      }
    }
    function createChannel(&$plugCfg, $label) { // FIXME: I need to receive new channel id here from the caller, which should do this atomically  
      if (count($this->channels) > 0) {
        $newId = max(array_keys($this->channels)) + 1;
      } else {
        $newId = 0;
      };
      $ret = new Channel($newId, $label);
      $this->channels[$newId] = $ret;
      $plugCfg[$newId] = $ret->save();
      return $ret;
    }
    function validateRule($ruleDef) {  
      // FIXME: rule condition could be an alias of the other rule in other channel , need to support verifying that
      // FIXME: need to actually request server to parse the rule for validity
      if (!$ruleDef || !is_array($ruleDef) || !isset($ruleDef["src"]) || !isset($ruleDef["cond"])) {
        return false;
      };
      $srcChId = intval($ruleDef["src"]);
      if (($srcChId < 0) || !isset($this->channels[$srcChId])) {
        return false;
      };
      if (!strlen(trim($ruleDef["cond"]))) {
        return false;
      };
      return true;
    }
    function canDeleteRule($chId, $ruleId) {
      return true; // FIXME: when we have references to rules 
    }
  }
?>
