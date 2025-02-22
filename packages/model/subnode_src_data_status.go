/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package model

type SubNodeSrcDataStatus struct {
	ID       int64  `gorm:"primary_key; not null" json:"id"`
	DataUUID string `gorm:"not null" json:"data_uuid"`
	TaskUUID string `gorm:"not null" json:"task_uuid"`
	Hash     string `gorm:"not null" json:"hash"`
	Data     []byte `gorm:"column:data;not null" json:"data"`
	DataInfo string `gorm:"type:jsonb" json:"data_info"`
	TranMode int64  `gorm:"not null" json:"tran_mode"`
	//SubNodeSrcPubkey     string `gorm:"not null" json:"subnode_src_pubkey"`
	SubNodeSrcPubkey string `gorm:"column:subnode_src_pubkey;not null" json:"subnode_src_pubkey"`
	//SubNodeDestPubkey    string `gorm:"not null" json:"subnode_dest_pubkey"`
	SubNodeDestPubkey string `gorm:"column:subnode_dest_pubkey;not null" json:"subnode_dest_pubkey"`
	//SubNodeDestIP        string `gorm:"not null" json:"subnode_dest_ip"`
	SubNodeDestIP string `gorm:"column:subnode_dest_ip;not null" json:"subnode_dest_ip"`
	//SubNodeAgentPubkey   string `gorm:"not null" json:"subnode_agent_pubkey"`
	SubNodeAgentPubkey string `gorm:"column:subnode_agent_pubkey;not null" json:"subnode_agent_pubkey"`
	//SubNodeAgentIP       string `gorm:"not null" json:"subnode_agent_ip"`
	SubNodeAgentIP string `gorm:"column:subnode_agent_ip;not null" json:"subnode_agent_ip"`
	AgentMode      int64  `gorm:"not null" json:"agent_mode"`
	DataSendState  int64  `gorm:"not null" json:"data_send_state"`
	DataSendErr    string `gorm:"not null" json:"data_send_err"`
	UpdateTime     int64  `gorm:"not null" json:"update_time"`
	CreateTime     int64  `gorm:"not null" json:"create_time"`
}

func (SubNodeSrcDataStatus) TableName() string {
	return "subnode_src_data_status"
}

func (m *SubNodeSrcDataStatus) Create() error {
	return DBConn.Create(&m).Error
}

func (m *SubNodeSrcDataStatus) Updates() error {
	return DBConn.Model(m).Updates(m).Error
}

func (m *SubNodeSrcDataStatus) Delete() error {
	return DBConn.Delete(m).Error
}

func (m *SubNodeSrcDataStatus) GetAll() ([]SubNodeSrcDataStatus, error) {
	var result []SubNodeSrcDataStatus
	err := DBConn.Find(&result).Error
	return result, err
}
func (m *SubNodeSrcDataStatus) GetOneByID() (*SubNodeSrcDataStatus, error) {
	err := DBConn.Where("id=?", m.ID).First(&m).Error
	return m, err
}

func (m *SubNodeSrcDataStatus) GetAllByTaskUUID(TaskUUID string) ([]SubNodeSrcDataStatus, error) {
	result := make([]SubNodeSrcDataStatus, 0)
	err := DBConn.Table("subnode_src_data_status").Where("task_uuid = ?", TaskUUID).Find(&result).Error
	return result, err
}

func (m *SubNodeSrcDataStatus) GetAllByDataSendStatus(DataSendStatus int64) ([]SubNodeSrcDataStatus, error) {
	result := make([]SubNodeSrcDataStatus, 0)
	err := DBConn.Table("subnode_src_data_status").Where("data_send_state = ?", DataSendStatus).Find(&result).Error
	return result, err
}

func (m *SubNodeSrcDataStatus) GetAllByDataSendStatusAndAgentMode(DataSendStatus int64, AgentMode int64) ([]SubNodeSrcDataStatus, error) {
	result := make([]SubNodeSrcDataStatus, 0)
	err := DBConn.Table("subnode_src_data_status").Where("data_send_state = ? AND agent_mode = ?", DataSendStatus, AgentMode).Find(&result).Error
	return result, err
}
