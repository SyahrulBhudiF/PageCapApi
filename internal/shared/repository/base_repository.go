package repository

import (
	"context"
	"fmt"
	_interface "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/repository/interface"

	"gorm.io/gorm"
)

type Repository[T any] struct {
	DB *gorm.DB
}

func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

func (r *Repository[T]) Delete(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Delete(entity).Error
}

func (r *Repository[T]) CountByUUID(ctx context.Context, uuid any) (int64, error) {
	var total int64
	err := r.DB.WithContext(ctx).Model(new(T)).Where("uuid = ?", uuid).Count(&total).Error
	return total, err
}

func (r *Repository[T]) FindByUUID(ctx context.Context, uuid any) (*T, error) {
	var entity T
	err := r.DB.WithContext(ctx).Where("uuid = ?", uuid).Take(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Repository[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.DB.WithContext(ctx).Find(&entities).Error
	return entities, err
}

func (r *Repository[T]) Find(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).First(entity, entity).Error
}

func (r *Repository[T]) Exists(ctx context.Context, entity *T) bool {
	return r.DB.WithContext(ctx).First(entity, entity).RowsAffected > 0
}

func (r *Repository[T]) FindByColumnValue(ctx context.Context, columnName string, value any) ([]T, error) {
	var entities []T
	err := r.DB.WithContext(ctx).Where(fmt.Sprintf("%s = ?", columnName), value).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *Repository[T]) BeginTx(ctx context.Context) _interface.IRepository[T] {
	return &Repository[T]{
		DB: r.DB.WithContext(ctx).Begin(),
	}
}

func (r *Repository[T]) Commit() error {
	return r.DB.Commit().Error
}

func (r *Repository[T]) Rollback() error {
	return r.DB.Rollback().Error
}

func (r *Repository[T]) Transaction(ctx context.Context, fn _interface.TransactionFunc[T]) error {
	tx := r.BeginTx(ctx)

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = fmt.Errorf("transaction error: %v, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	return tx.Commit()
}
