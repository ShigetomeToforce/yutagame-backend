package usecase

import (
	"context"
	"yutagame-backend/domain/model"
	"yutagame-backend/infrastructure/database"
)

type MachineUseCase struct {
	machineRepo *database.MachineRepository
}

func NewMachineUseCase(machineRepo *database.MachineRepository) *MachineUseCase {
	return &MachineUseCase{machineRepo: machineRepo}
}

func (u *MachineUseCase) GetAllMachines(ctx context.Context) ([]model.Machine, error) {
	return u.machineRepo.FindAll(ctx)
}

func (u *MachineUseCase) GetMachineByID(ctx context.Context, id int64) (*model.Machine, error) {
	return u.machineRepo.FindByID(ctx, id)
}

func (u *MachineUseCase) CreateMachine(ctx context.Context, m *model.Machine) error {
	return u.machineRepo.Create(ctx, m)
}

func (u *MachineUseCase) UpdateMachine(ctx context.Context, m *model.Machine) error {
	return u.machineRepo.Update(ctx, m)
}

func (u *MachineUseCase) DeleteMachine(ctx context.Context, id int64) error {
	return u.machineRepo.Delete(ctx, id)
}
