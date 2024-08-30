package main

import (
	"dagproc/internal/di"
	"sync"
	"github.com/rs/zerolog/log"

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
		RxCh chan di.DagMsg
		CtsCh chan di.ChanPlugCtx
		opcs map[di.ChanPlugCtx]di.PlugCommPtr // output plugin communications slice
		chpt di.ChannelPtr // channel pointer
		chplct di.ChanPlugCtx // channel plugin context
		plcopt di.PlugCommPtr // plugin communication 
		ok bool
	)
	RxCh = make(chan di.DagMsg)
	CtsCh = make(chan di.ChanPlugCtx)
	opcs = make(map[di.ChanPlugCtx]di.PlugCommPtr)
	for _, chpt = range(cfg.Channels) { // channel name, channel pointers
		for _, chplct = range(chpt.InPlugs) { // channelpluginctx
			wg.Add(1)
			go inGoro(chpt, chplct, RxCh, &wg)
		}
		for _, chplct = range(chpt.OutPlugs) {
			wg.Add(1)
			plcopt = new (di.PlugComm)
			plcopt.TxChan = make(chan di.DagMsg)
			plcopt.CTS = false // should be raised on OutGoroHook
			plcopt.Buffer = make([]di.DagMsg, 0)
			opcs[chplct] = plcopt // Each output plugin context needs it's separate goroutine, and respective plugin communication context
			go outGoro(chpt, chplct, plcopt.TxChan, CtsCh, &wg)
		}

	} 
	// At this point we have as many goroutines as there are input and output plugin contexts.
	// Now, magic
	for (true) {
		var (
			inMsg di.DagMsg
			ctsMsg di.ChanPlugCtx
			outMsgs []di.DagMsg
			om di.DagMsg
		)
		select {
			case inMsg = <- RxCh:
				// Got a message from the input plugin, now we have to process it
				outMsgs = processInMsg(inMsg) // Should return a slice of DagMsg's with messages to send and channels to send those messages to
				for _, om = range(outMsgs) {
					for _, chplct = range(om.Channel.OutPlugs) {
						plcopt,ok = opcs[chplct]
						if (!ok) {
							log.Info().Msgf("Unknown log level %s, actual log level set to info", logLevel)
						}
						if opcs[chplct].CTS {
							opcs[chplct].TxChan <- om
						} else {
							opcs[chplct].Buffer = append(opcs[chplct].Buffer, om) // rpush
						}
					}
				}
			case ctsMsg = <- CtsCh:
				// Plugin reports it is Clear To Send
				if len(opcs[cpc].Buffer) > 0 {
					// lpop
					opcs[cpc].TxChan <- opcs[cpc].Buffer[0:1]  
					opcs[cpc].Buffer = opcs[cpc].Buffer[1:] 
				} else {
					opcs[cpc].CTS = true
				}
		}
	}
	wg.Wait()
}

func ProcessInMsg(msg di.DagMsg) []di.DagMsg {
	var (
		ret []di.DagMsg
	)
	ret = make([]di.DagMsg, 0)
	return ret
}

func inGoro(ChPt di.ChannelPtr, ChPCtx di.ChanPlugCtx, RxChan chan di.DagMsg, wg *sync.WaitGroup) error {
	var (
		ret error
	)
	wg.Done()
	return ret
} 

func outGoro(ChPt di.ChannelPtr, ChPCtx di.ChanPlugCtx, TxChan chan di.DagMsg, CtsChan chan di.ChanPlugCtx , wg *sync.WaitGroup) error {
	var (
		ret error
	)
	wg.Done()
	return ret
}

