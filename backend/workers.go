package main

import (
	"dagproc/internal/di"
	"sync"
)

/*
	the idea behind this function is:
	Golang select semantics does not allow listening on arbitrary set of channels, unlike real select(2) , so we're multiplexing input plugins on one shared
	chan (inCh); after receiving DagMsgPtr from any of the input plugins we're processing received data using dm_core.RunTheGauntlet; output of this function is
	the plugin name this message should be sent to ; to avoid using non blocking select {} we're requiring plugins to report when they are ready to send via
	second multiplexed chah (outInch). Upon receiving anything from that chan we know this particular plugin is free to process another DagMsg, so we're raising
	respective CTS flag . If this flag is lowered ,we're buffering the message until this flag is raised.
*/
func runWorkers(cfg di.GlobalConfig) {
	var (
		wg sync.WaitGroup
		RxCh, CtsCh chan di.DagMsgPtr
		pcs map[di.PluginPtr]di.PlugComm
		pn string
		pp di.PluginPtr
	)
	RxCh = make(chan di.DagMsgPtr)
	CtsCh = make(chan di.DagMsgPtr)
	pcs = make(map[di.PluginPtr]di.PlugComm)
	wg.Add(len(plugins)) 
	for chn, chp = range(cfg.Channels) { // channel name, channel points
		for cpc = range(chp[InPlugs]) { // channelpluginctx

		}
	}
	for pn, pp = range(cfg.Plugins) {
		pcs[pp].Channels = dm_core.ChannelsUsingPlugin(pn, cfg.Channels)
		if ((pp.Type & di.PT_IN) != 0) {
			go pp.Module.Hooks.InGoroHook(pp, RxCh, &wg)
		} else 
		if ((pp.Type & di.PT_OUT) != 0) {
			pcs[pp].TxCh = make(chan di.DagMsgPtr)
			pcs[pp].CTS = false // Should get cleared after OutGoroHook runs 
			pcs[pp].SendBuffer = make([]DagMsgPtr)
			go pp.Module.Hooks.OutGoroHook(pp, pcs[pp].TxCh, СtsCh, &wg)
		} else {
			wg.Done()
		}
	}
	// Now, magic
	while (true) {
		var (
			inMsg, ctsMsg DagMsgPtr
			outPlugs []di.PluginPtr
			op di.PluginPtr
		)
		select {
			case inMsg 	<- RxCh:
				// Got a message from the input plugin, now we have to process it
				outPlugs = dm_core.ProcessMsg(inMsg, plugins)
				for op = range(outPlugs) {
					if pcs[op].CTS {
						pcs[op].TxCh <- inMsg
					} else {
						pcs[op].SendBuffer = append(pcs[op].SendBuffer, inMsg) // rpush
					}
				}
			case ctsMsg <- CtsCh:
				// Plugin reports it is Clear To Send
				pp = ctsMsg.Plugin
				if len(pcs[pp].SendBuffer) > 0 {
					// lpop
					pcs[pp].TxCh <- pcs[pp].SendBuffer[0:1]  
					pcs[pp].SendBuffer = pcs[pp].SendBuffer[1:] 
				} else {
					pcs[pp].CTS = true
				}
		}
	}
	wg.Wait()
}
