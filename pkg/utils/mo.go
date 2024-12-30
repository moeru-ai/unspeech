package utils

import (
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func FilterOptionPresent[T any](item mo.Option[T], _ int) bool {
	return item.IsPresent()
}

func FilterOptionAbsent[T any](item mo.Option[T], _ int) bool {
	return item.IsAbsent()
}

func MapOptionOrEmpty[T any](item mo.Option[T], _ int) T {
	return item.OrEmpty()
}

func MapOptionMust(item mo.Option[error], _ int) error {
	return item.MustGet()
}

func MapOptionsPresent[T any](items []mo.Option[T]) []T {
	filtered := lo.Filter(items, FilterOptionPresent)
	return lo.Map(filtered, MapOptionOrEmpty)
}

func ResultToOption[T any](item mo.Result[T]) mo.Option[T] {
	if item.IsError() {
		return mo.None[T]()
	}

	return mo.Some(item.MustGet())
}
