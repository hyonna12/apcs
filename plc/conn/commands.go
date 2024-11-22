package conn

import (
	"apcs_refactored/config"
	"fmt"
)

func SendRobotMove(robotId int, position int, commandId string) error {
	var robotSpeed int
	for _, robot := range config.Config.Plc.Resource.Robot.Robots {
		if robot.ID == robotId {
			robotSpeed = robot.Speed
			break
		}
	}

	cmd := &PLCCommand{
		Type: "robot_move",
		Params: map[string]interface{}{
			"robot_id":   robotId,
			"position":   position,
			"speed":      robotSpeed,
			"command_id": commandId,
		},
	}

	resp, err := client.SendCommand(cmd)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}
	return nil
}

func SendRobotRotate(robotId int, direction string, commandId string) error {
	cmd := &PLCCommand{
		Type: "robot_rotate",
		Params: map[string]interface{}{
			"robot_id":   robotId,
			"direction":  direction,
			"command_id": commandId,
		},
	}

	resp, err := client.SendCommand(cmd)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}
	return nil
}

func SendRobotHandler(robotId int, operation string, commandId string) error {
	cmd := &PLCCommand{
		Type: "robot_handler",
		Params: map[string]interface{}{
			"robot_id":   robotId,
			"operation":  operation,
			"command_id": commandId,
		},
	}

	resp, err := client.SendCommand(cmd)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}
	return nil
}

func SendDoorOperation(doorType string, operation string, commandId string) error {
	cmd := &PLCCommand{
		Type: "door_operation",
		Params: map[string]interface{}{
			"door_type":  doorType,
			"operation":  operation,
			"command_id": commandId,
		},
	}

	resp, err := client.SendCommand(cmd)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}
	return nil
}

func SendTrayBufferOperation(operation string, commandId string) error {
	cmd := &PLCCommand{
		Type: "tray_buffer",
		Params: map[string]interface{}{
			"operation":  operation,
			"command_id": commandId,
		},
	}

	resp, err := client.SendCommand(cmd)
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}
	return nil
}
