package domain

type AccessController interface {
}

const (
	AccessRoleAdmin = iota
	AccessRoleOperator
	AccessRoleGuest
	AccessRoleHost
)

const (
	AccessObjectOrder = iota
	AccessObjectHousing
	AccessObjectDwelling
	AccessObjectLot
)
