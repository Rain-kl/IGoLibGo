// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"Wavelet/core/contracts"
)

// Note: 命令名/别名全局唯一、自带指令走 RegisterBuiltin — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

// CommandMeta is the listing view of a registered bot command.
type CommandMeta struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Builtin     bool
}

type registeredCommand struct {
	cmd          contracts.BotCommand
	meta         CommandMeta
	allowUnbound bool
}

// BotRegistry is the in-process command and conversation table.
// Register is startup-only; Lookup is safe for concurrent inbound dispatch.
type BotRegistry struct {
	mu       sync.RWMutex
	occupied map[string]*registeredCommand
	commands []*registeredCommand
	convs    map[string]contracts.BotConversation
}

var _ contracts.BotCommandRegistry = (*BotRegistry)(nil)

var (
	botRegMu sync.RWMutex
	botReg   *BotRegistry
)

// SetBotRegistry stores the process-level command registry used by the runner.
func SetBotRegistry(r *BotRegistry) {
	botRegMu.Lock()
	defer botRegMu.Unlock()
	botReg = r
}

// GetBotRegistry returns the process-level command registry, or nil.
func GetBotRegistry() *BotRegistry {
	botRegMu.RLock()
	defer botRegMu.RUnlock()
	return botReg
}

// NewBotRegistry creates an empty command registry.
func NewBotRegistry() *BotRegistry {
	return &BotRegistry{
		occupied: make(map[string]*registeredCommand),
		convs:    make(map[string]contracts.BotConversation),
	}
}

// Register adds a business command. Business commands require a bound user.
func (r *BotRegistry) Register(cmd contracts.BotCommand) error {
	return r.register(cmd, false, false)
}

// RegisterBuiltin adds a gateway builtin command. It is not on contracts.BotCommandRegistry.
func (r *BotRegistry) RegisterBuiltin(cmd contracts.BotCommand, allowUnbound bool) error {
	return r.register(cmd, true, allowUnbound)
}

func (r *BotRegistry) register(cmd contracts.BotCommand, builtin, allowUnbound bool) error {
	if cmd == nil {
		return fmt.Errorf("bot command is nil")
	}

	name := strings.TrimSpace(cmd.Name())
	desc := strings.TrimSpace(cmd.Description())
	usage := strings.TrimSpace(cmd.Usage())
	if name == "" {
		return fmt.Errorf("bot command name is empty")
	}
	if desc == "" {
		return fmt.Errorf("bot command %q description is empty", name)
	}
	if usage == "" {
		return fmt.Errorf("bot command %q usage is empty", name)
	}
	if err := validateCommandToken(name); err != nil {
		return err
	}

	keys := make([]string, 0, 1+len(cmd.Aliases()))
	seen := make(map[string]struct{}, 1+len(cmd.Aliases()))
	addKey := func(raw string) error {
		token := strings.TrimSpace(raw)
		if token == "" {
			return fmt.Errorf("bot command %q has empty alias", name)
		}
		if err := validateCommandToken(token); err != nil {
			return err
		}
		key := normalizeCommandName(token)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("bot command name %q already registered", key)
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
		return nil
	}
	if err := addKey(name); err != nil {
		return err
	}
	aliases := cmd.Aliases()
	copiedAliases := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		if err := addKey(alias); err != nil {
			return err
		}
		copiedAliases = append(copiedAliases, strings.TrimSpace(alias))
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, key := range keys {
		if _, ok := r.occupied[key]; ok {
			return fmt.Errorf("bot command name %q already registered", key)
		}
	}

	rec := &registeredCommand{
		cmd: cmd,
		meta: CommandMeta{
			Name:        name,
			Aliases:     copiedAliases,
			Description: desc,
			Usage:       usage,
			Builtin:     builtin,
		},
		allowUnbound: allowUnbound,
	}
	for _, key := range keys {
		r.occupied[key] = rec
	}
	r.commands = append(r.commands, rec)
	return nil
}

// RegisterConversation adds a multi-turn handler. Names must contain '.' and are unique.
func (r *BotRegistry) RegisterConversation(conv contracts.BotConversation) error {
	if conv == nil {
		return fmt.Errorf("bot conversation is nil")
	}
	raw := strings.TrimSpace(conv.Name())
	name := strings.ToLower(raw)
	if !strings.Contains(name, ".") {
		return fmt.Errorf("bot conversation name %q must contain '.'", raw)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.convs[name]; ok {
		return fmt.Errorf("bot conversation %q already registered", name)
	}
	r.convs[name] = conv
	return nil
}

// LookupCommand finds a command by name or alias. Matching is case-insensitive.
func (r *BotRegistry) LookupCommand(name string) (cmd contracts.BotCommand, builtin, allowUnbound, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.occupied[normalizeCommandName(name)]
	if !ok {
		return nil, false, false, false
	}
	return rec.cmd, rec.meta.Builtin, rec.allowUnbound, true
}

// LookupConversation finds a conversation by name. Matching is case-insensitive.
func (r *BotRegistry) LookupConversation(name string) (contracts.BotConversation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conv, ok := r.convs[strings.ToLower(strings.TrimSpace(name))]
	return conv, ok
}

// List returns registered commands in registration order.
func (r *BotRegistry) List() []CommandMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]CommandMeta, len(r.commands))
	for i, rec := range r.commands {
		meta := rec.meta
		if rec.meta.Aliases != nil {
			meta.Aliases = append([]string(nil), rec.meta.Aliases...)
		}
		out[i] = meta
	}
	return out
}

func normalizeCommandName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func validateCommandToken(name string) error {
	for _, r := range name {
		if r == '/' || unicode.IsSpace(r) {
			return fmt.Errorf("bot command name %q contains '/' or whitespace", name)
		}
	}
	return nil
}
