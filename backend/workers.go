package main

import (
	"dagproc/internal/di"
	"dagproc/internal/dm_core"
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
		wg          sync.WaitGroup
		RxCh        chan di.DagMsgPtr
		CtsCh       chan di.ChanPlugCtxPtr
		opcs        map[di.ChanPlugCtxPtr]di.PlugCommPtr // output plugin communications slice
		chpt        di.ChannelPtr                        // channel pointer
		chplct      di.ChanPlugCtxPtr                    // channel plugin context
		plcopt      di.PlugCommPtr                       // plugin communication
		ok          bool
		lastEventId uint64
	)
	RxCh = make(chan di.DagMsgPtr)
	CtsCh = make(chan di.ChanPlugCtxPtr)
	opcs = make(map[di.ChanPlugCtxPtr]di.PlugCommPtr)
	lastEventId = 0
	for _, chpt = range cfg.Channels { // channel name, channel pointers
		for _, chplct = range chpt.InPlugs { // channelpluginctxptr
			wg.Add(1)
			go inGoro(chpt, chplct, RxCh, &wg)
		}
		for _, chplct = range chpt.OutPlugs {
			wg.Add(1)
			plcopt = &di.PlugComm{
				TxChan: make(chan di.DagMsgPtr),
				CTS:    false, // should get raised in OutGoro
				Buffer: make([]di.DagMsgPtr, 0),
			}
			opcs[chplct] = plcopt // Each output plugin context needs it's separate goroutine, and respective plugin communication context
			go outGoro(chplct, plcopt.TxChan, CtsCh, &wg)
		}

	}
	// At this point we have as many goroutines as there are input and output plugin contexts.
	// Now, magic
	for true {
		var (
			inMsg   di.DagMsgPtr
			ctsMsg  di.ChanPlugCtxPtr
			outMsgs []di.DagMsgPtr
			om      di.DagMsgPtr
			err     error
		)
		select {
		case inMsg = <-RxCh:
			// Got a message from the input plugin, now we have to process it
			inMsg.Id = lastEventId
			lastEventId = lastEventId + 1

			outMsgs, err = processInMsg(inMsg) // Should return a slice of DagMsg's with messages to send and channels to send those messages to
			if err != nil {
				log.Error().Uint64("Event id", inMsg.Id).Err(err).Msg("Error while processing event, skipping")
				continue
			}
			for _, om = range outMsgs {
				for _, chplct = range om.Channel.OutPlugs {
					plcopt, ok = opcs[chplct]
					if !ok {
						log.Error().Msgf("Request to send message to the unknown plugin of channel %s", om.Channel.Name)
						continue
					}
					if plcopt.CTS {
						plcopt.TxChan <- om
						plcopt.CTS = false
					} else {
						plcopt.Buffer = append(opcs[chplct].Buffer, om) // rpush
					}
				}
			}
		case ctsMsg = <-CtsCh:
			// Plugin reports it is Clear To Send
			plcopt, ok = opcs[ctsMsg]
			if !ok {
				log.Error().Msgf("Plugin %s reports readiness for an unknown context %v", ctsMsg.Plugin.Name, ctsMsg)
				continue
			}
			if len(plcopt.Buffer) > 0 {
				// lpop
				plcopt.TxChan <- plcopt.Buffer[0]
				plcopt.Buffer = plcopt.Buffer[1:]
			} else {
				plcopt.CTS = true
			}
		}
	}
	wg.Wait()
}

func processInMsg(msg di.DagMsgPtr) ([]di.DagMsgPtr, error) {
	var (
		ret []di.DagMsgPtr
	)
	ret = make([]di.DagMsgPtr, 0)
	ret = dm_core.ChannelMatchAndFlush(msg.Channel, msg.Channel, msg, ret)
	return ret, nil
}

func inGoro(ch di.ChannelPtr, chplct di.ChanPlugCtxPtr, rxChan chan di.DagMsgPtr, wg *sync.WaitGroup) error {
	var (
		err error
		nm  di.DagMsgPtr
	)
	for true {
		nm, err = chplct.Plugin.Module.Hooks.ReceiveMsgHook(chplct)
		if err != nil {
			// some error up the stack, resume the loop
			continue
		}
		nm.Channel = ch
		rxChan <- nm
	}
	wg.Done()
	return err
}

func outGoro(chplct di.ChanPlugCtxPtr, txChan chan di.DagMsgPtr, ctsChan chan di.ChanPlugCtxPtr, wg *sync.WaitGroup) error {
	var (
		err error
		nm  di.DagMsgPtr
	)
	for true {
		ctsChan <- chplct // Clearing CTS flag in main thread
		nm = <-txChan
		err = chplct.Plugin.Module.Hooks.SendMsgHook(nm, chplct)
		if err != nil {
			// Error delivering the message, drop it for now (FIXME)
		}
	}
	wg.Done()
	return err
}
