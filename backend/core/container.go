// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package core provides the micro-kernel service bus, generic IoC container, and runtime extensions.
package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Container manages service registration and resolution using Go reflection and generics.
//
// Note: Go 1.27 特性与现代语法糖重构见 .agents/notes/implemented/simplification/2026-09-09-go127-modernization.md
type Container struct {
	mu             sync.RWMutex
	parent         *Container
	services       map[reflect.Type]any
	interfaceCache map[reflect.Type]any
	listeners      map[reflect.Type][]func(any)
}

// NewContainer creates a new IoC container instance with an optional parent container.
func NewContainer(parent *Container) *Container {
	return &Container{
		parent:         parent,
		services:       make(map[reflect.Type]any),
		interfaceCache: make(map[reflect.Type]any),
		listeners:      make(map[reflect.Type][]func(any)),
	}
}

func isNil(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func (c *Container) remove(targetType reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.services, targetType)
	clear(c.interfaceCache)
}

// Provide registers a typed service implementation into the Context hierarchy's root IoC container.
func Provide[T any](ctx *Context, service T) {
	if ctx == nil {
		panic("core: nil context provided to Provide")
	}
	if isNil(service) {
		panic("core: cannot provide nil service")
	}

	targetType := reflect.TypeFor[T]()
	targetContainer := ctx.Root().Container()
	targetContainer.provide(targetType, service)

	ctx.OnDispose(func() error {
		targetContainer.remove(targetType)
		return nil
	})
}

// ProvideScoped registers a typed service implementation strictly in the local Context container.
func ProvideScoped[T any](ctx *Context, service T) {
	if ctx == nil {
		panic("core: nil context provided to ProvideScoped")
	}
	if isNil(service) {
		panic("core: cannot provide nil service")
	}

	targetType := reflect.TypeFor[T]()
	targetContainer := ctx.Container()
	targetContainer.provide(targetType, service)

	ctx.OnDispose(func() error {
		targetContainer.remove(targetType)
		return nil
	})
}

func (c *Container) provide(targetType reflect.Type, service any) {
	c.mu.Lock()
	c.services[targetType] = service
	clear(c.interfaceCache)

	// Collect any matching listeners to invoke outside the lock
	var callbacks []func(any)
	svcType := reflect.TypeOf(service)
	for lType, cbs := range c.listeners {
		if lType == targetType || (lType.Kind() == reflect.Interface && svcType.Implements(lType)) {
			callbacks = append(callbacks, cbs...)
		}
	}
	c.mu.Unlock()

	for _, cb := range callbacks {
		cb(service)
	}
}

// Inject resolves a registered service of type T from the Context.
func Inject[T any](ctx *Context) (T, error) {
	var zero T
	if ctx == nil {
		return zero, ErrNilContext
	}

	targetType := reflect.TypeFor[T]()
	val, err := ctx.Container().resolve(targetType)
	if err != nil {
		return zero, err
	}

	typedVal, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("%w: cannot cast %T to %v", ErrServiceNotFound, val, targetType)
	}
	return typedVal, nil
}

func (c *Container) resolve(targetType reflect.Type) (any, error) {
	c.mu.RLock()
	// 1. Direct type match
	if val, ok := c.services[targetType]; ok {
		c.mu.RUnlock()
		return val, nil
	}
	c.mu.RUnlock()

	// 2. Interface assignment scan & cache
	if targetType.Kind() == reflect.Interface {
		if val, found := c.resolveInterface(targetType); found {
			return val, nil
		}
	}

	// 3. Fallback to parent container
	if c.parent != nil {
		return c.parent.resolve(targetType)
	}

	return nil, fmt.Errorf("%w: %v", ErrServiceNotFound, targetType)
}

func (c *Container) resolveInterface(targetType reflect.Type) (any, bool) {
	c.mu.RLock()
	if val, ok := c.interfaceCache[targetType]; ok {
		c.mu.RUnlock()
		return val, true
	}

	var matched any
	for _, val := range c.services {
		if reflect.TypeOf(val).Implements(targetType) {
			matched = val
			break
		}
	}
	c.mu.RUnlock()

	if matched == nil {
		return nil, false
	}

	c.mu.Lock()
	if c.interfaceCache == nil {
		c.interfaceCache = make(map[reflect.Type]any)
	}
	c.interfaceCache[targetType] = matched
	c.mu.Unlock()
	return matched, true
}

// MustInject resolves a service of type T or panics if the service is not found.
func MustInject[T any](ctx *Context) T {
	s, err := Inject[T](ctx)
	if err != nil {
		panic(fmt.Sprintf("core: failed to inject service %v: %v", reflect.TypeFor[T](), err))
	}
	return s
}

// Has returns true if a service of type T is registered and resolvable in the Context.
func Has[T any](ctx *Context) bool {
	_, err := Inject[T](ctx)
	return err == nil
}

// Using executes the given function synchronously if the required dependency is ready.
func Using[T1 any](ctx *Context, fn func(s1 T1)) error {
	s1, err := Inject[T1](ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrServiceNotReady, err)
	}
	fn(s1)
	return nil
}

// Using2 executes the given function synchronously if both required dependencies are ready.
func Using2[T1, T2 any](ctx *Context, fn func(s1 T1, s2 T2)) error {
	s1, err1 := Inject[T1](ctx)
	s2, err2 := Inject[T2](ctx)
	if err := errors.Join(err1, err2); err != nil {
		return fmt.Errorf("%w: %w", ErrServiceNotReady, err)
	}
	fn(s1, s2)
	return nil
}

// Using3 executes the given function synchronously if all 3 required dependencies are ready.
func Using3[T1, T2, T3 any](ctx *Context, fn func(s1 T1, s2 T2, s3 T3)) error {
	s1, err1 := Inject[T1](ctx)
	s2, err2 := Inject[T2](ctx)
	s3, err3 := Inject[T3](ctx)
	if err := errors.Join(err1, err2, err3); err != nil {
		return fmt.Errorf("%w: %w", ErrServiceNotReady, err)
	}
	fn(s1, s2, s3)
	return nil
}

// When registers a reactive hook that is called immediately if T is already provided,
// or called as soon as T is provided in the future.
//
// Listeners are stored on the root container so they observe core.Provide, which
// always writes to the root. Registering on a Fiber child container would miss
// services provided by plugins that load later.
func When[T any](ctx *Context, fn func(s T)) {
	if ctx == nil {
		panic("core: nil context provided to When")
	}

	targetType := reflect.TypeFor[T]()
	c := ctx.Root().Container()

	// If already ready, execute immediately
	if s, err := Inject[T](ctx); err == nil {
		fn(s)
	}

	// Also register listener for future calls / updates
	c.mu.Lock()
	defer c.mu.Unlock()
	c.listeners[targetType] = append(c.listeners[targetType], func(val any) {
		if typed, ok := val.(T); ok {
			fn(typed)
		}
	})
}

// Bind is When with a name that matches plugin wiring: fill a dependency as
// soon as the root container provides it.
func Bind[T any](ctx *Context, fn func(s T)) {
	When(ctx, fn)
}

// ============================================================================
// Go 1.27 结构体泛型方法 (Generic Methods on *Context & *Container)
// ============================================================================

// Provide registers a typed service implementation into the Context hierarchy's root IoC container.
func (c *Context) Provide[T any](service T) {
	Provide[T](c, service)
}

// ProvideScoped registers a typed service implementation strictly in the local Context container.
func (c *Context) ProvideScoped[T any](service T) {
	ProvideScoped[T](c, service)
}

// Inject resolves a registered service of type T from this Context.
func (c *Context) Inject[T any]() (T, error) {
	return Inject[T](c)
}

// MustInject resolves a service of type T from this Context or panics if not found.
func (c *Context) MustInject[T any]() T {
	return MustInject[T](c)
}

// Has returns true if a service of type T is registered and resolvable in this Context.
func (c *Context) Has[T any]() bool {
	return Has[T](c)
}

// Using executes the given function synchronously if the required dependency is ready.
func (c *Context) Using[T1 any](fn func(s1 T1)) error {
	return Using[T1](c, fn)
}

// Using2 executes the given function synchronously if both required dependencies are ready.
func (c *Context) Using2[T1, T2 any](fn func(s1 T1, s2 T2)) error {
	return Using2[T1, T2](c, fn)
}

// Using3 executes the given function synchronously if all 3 required dependencies are ready.
func (c *Context) Using3[T1, T2, T3 any](fn func(s1 T1, s2 T2, s3 T3)) error {
	return Using3[T1, T2, T3](c, fn)
}

// When registers a reactive hook that is called immediately if T is already provided,
// or called as soon as T is provided in the future.
func (c *Context) When[T any](fn func(s T)) {
	When[T](c, fn)
}

// Bind is When with a name that matches plugin wiring: fill a dependency as
// soon as the root container provides it.
func (c *Context) Bind[T any](fn func(s T)) {
	Bind[T](c, fn)
}

// Provide registers a typed service implementation into this Container.
func (c *Container) Provide[T any](service T) {
	if c == nil {
		panic("core: nil container provided to Provide")
	}
	if isNil(service) {
		panic("core: cannot provide nil service")
	}
	targetType := reflect.TypeFor[T]()
	c.provide(targetType, service)
}

// Inject resolves a registered service of type T from this Container.
func (c *Container) Inject[T any]() (T, error) {
	var zero T
	if c == nil {
		return zero, ErrNilContainer
	}
	targetType := reflect.TypeFor[T]()
	val, err := c.resolve(targetType)
	if err != nil {
		return zero, err
	}
	typedVal, ok := val.(T)
	if !ok {
		return zero, fmt.Errorf("%w: cannot cast %T to %v", ErrServiceNotFound, val, targetType)
	}
	return typedVal, nil
}

// MustInject resolves a service of type T from this Container or panics if not found.
func (c *Container) MustInject[T any]() T {
	s, err := c.Inject[T]()
	if err != nil {
		panic(fmt.Sprintf("core: failed to inject service %v: %v", reflect.TypeFor[T](), err))
	}
	return s
}

// Has returns true if a service of type T is registered and resolvable in this Container.
func (c *Container) Has[T any]() bool {
	_, err := c.Inject[T]()
	return err == nil
}
