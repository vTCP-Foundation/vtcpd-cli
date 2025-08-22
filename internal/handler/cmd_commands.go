package handler

import (
	"errors"
	"fmt"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (nh *NodeHandler) HandleStart() error {
	isNodeRunning, err := nh.CheckNodeRunning()
	if err != nil {
		logger.Error("Can't check if node is running. Details: " + err.Error())
		return err
	}
	if isNodeRunning {
		logger.Error("Node already running")
		return errors.New("Node already running")
	}
	err = nh.RestoreNode()
	if err != nil {
		logger.Error("Node can't be restored. Error details: " + err.Error())
		return errors.New("Can't start. " + err.Error())
	}
	fmt.Println("Started")
	return nil
}

func (nh *NodeHandler) HandleStop() error {
	err := nh.StopNode()
	if err != nil {
		logger.Error("Can't stop node " + err.Error())
		return errors.New("Can't stop node " + err.Error())
	}
	logger.Info("Node stopped")
	fmt.Println("Stopped")
	return nil
}

func (nh *NodeHandler) HandleChannels() error {
	err := nh.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		return errors.New("Node is not running. Details: " + err.Error())
	}
	nh.Channels()
	return nil
}

func (nh *NodeHandler) HandleSettlementLines() error {
	err := nh.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		return errors.New("Node is not running. Details: " + err.Error())
	}
	nh.SettlementLines()
	return nil
}

func (nh *NodeHandler) HandleMaxFlow() error {
	err := nh.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		return errors.New("Node is not running. Details: " + err.Error())
	}
	nh.MaxFlow()
	return nil
}

func (nh *NodeHandler) HandlePayment() error {
	err := nh.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		return errors.New("Node is not running. Details: " + err.Error())
	}
	nh.Payment()
	return nil
}

func (nh *NodeHandler) HandleHistory() error {
	err := nh.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		return errors.New("Node is not running. Details: " + err.Error())
	}
	nh.History()
	return nil
}

func (nh *NodeHandler) HandleRemoveOutdatedCrypto() error {
	err := nh.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		return errors.New("Node is not running. Details: " + err.Error())
	}
	nh.RemoveOutdatedCryptoDataCommand()
	return nil
}
