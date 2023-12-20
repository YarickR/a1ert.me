<?php
namespace Yjr\A1ertme;
  define("MLOCK", "mlock");
  class Channel {
    public int $id ;
    public string $label;
    public mixed $format = false;
    public array $rules = [];
    public array $sources = [];
    public array $sinks = [];
    function __construct($id = -1, $label  = "", $format = false, $rules = []) {
      $this->id = $id;
      $this->label = $label;
      $this->format = $format;
      $this->rules = $rules;
    }
    
    function init($ch) {
      $props = ["id" => true, "label" => true, "rules" => false, "format" => false];
      foreach ($props as $p => $m) {
        if (!isset($ch[$p])) {
          if ($m) {
            Logger::log(DEBUG, $ch);
            Logger::log(ERR, "Channel ", $ch, " missing mandatory property ", $p); 
            $this->id = -1;
            break;
          }
        } else {
          $this->{$p} = $p == "id" ? intval($ch[$p]) : $ch[$p];
        };
      };
      return $this;
    }

    function load($cfgField) {
      $chDef = json_decode($cfgField, true); 
        if ($chDef == null || !is_array($chDef) || (count($chDef) < 1)) {
        return false;
      };
      $this->init($chDef);
      return $this->id >= 0 ;
    }

    function save() {
      if ($this->id < 0) {
        Logger::log(DEBUG, $this);
        Logger::log(ERR, "Refusing to save uninitialised channel");
        return false;
      };
      $props = ['id' => true, 'label' => true, 'format' => false, 'rules' => false];
      $he = [];
      foreach ($props as $p => $m) {
        $he[$p] = $this->{$p};
      };
      return json_encode($he);
    }
    function createRule($rule) {
      $rIds = is_array($this->rules) ? array_keys($this->rules) : [];
      $newRid = sizeof($rIds) > 0 ? max($rIds) + 1 : 0;
      $rule["id"] = $newRid;
      $this->rules[$newRid] = $rule;
      return true;
    }

    function deleteRule($ruleId) {
      if (isset($this->rules[$ruleId])) {
        unset($this->rules[$ruleId]);
        return true;
      };
      return false;
    }
  }

?>