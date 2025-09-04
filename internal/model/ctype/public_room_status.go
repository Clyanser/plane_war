package ctype

type RoomStatus int8

const (
	Waiting RoomStatus = iota + 1
	Playing
)
