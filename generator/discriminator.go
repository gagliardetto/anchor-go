package generator

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
)

func (g *Generator) gen_discriminators() (*OutputFile, error) {
	file := NewFile(g.options.Package)
	file.HeaderComment("Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.")
	file.HeaderComment("This file contains the discriminators for accounts and events defined in the IDL.")

	{
		accountDiscriminatorsCodes := Empty()
		accountDiscriminatorsCodes.Comment("Account discriminators")
		accountDiscriminatorsCodes.Line()
		accountDiscriminatorsCodes.Var().Parens(
			DoGroup(func(code *Group) {
				for _, account := range g.idl.Accounts {
					if account.Discriminator == nil {
						continue
					}

					discriminator := account.Discriminator
					if len(discriminator) != 8 {
						panic(fmt.Errorf("discriminator for account %s must be exactly 8 bytes long, got %d bytes", account.Name, len(discriminator)))
					}

					discriminatorName := FormatAccountDiscriminatorName(account.Name)
					{
						code.Id(discriminatorName).Op("=").Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
							for _, byteVal := range discriminator[:] {
								byteGroup.Lit(int(byteVal))
							}
						}).Op("}")
					}
					code.Line()
				}
			}),
		)
		file.Add(accountDiscriminatorsCodes)
		file.Line()
	}
	{
		// Generate the discriminators for events.
		eventDiscriminatorsCodes := Empty()
		eventDiscriminatorsCodes.Comment("Event discriminators")
		eventDiscriminatorsCodes.Line()
		eventDiscriminatorsCodes.Var().Parens(
			DoGroup(func(code *Group) {
				for _, event := range g.idl.Events {
					if event.Discriminator == nil {
						continue
					}

					discriminator := event.Discriminator
					if len(discriminator) != 8 {
						panic(fmt.Errorf("discriminator for event %s must be exactly 8 bytes long", event.Name))
					}

					discriminatorName := FormatEventDiscriminatorName(event.Name)
					{
						code.Id(discriminatorName).Op("=").Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
							for _, byteVal := range discriminator[:] {
								byteGroup.Lit(int(byteVal))
							}
						}).Op("}")
					}
					code.Line()
				}
			}),
		)
		file.Add(eventDiscriminatorsCodes)
		file.Line()
	}
	{
		// Generate the discriminators for instructions.
		instructionDiscriminatorsCodes := Empty()
		instructionDiscriminatorsCodes.Comment("Instruction discriminators")
		instructionDiscriminatorsCodes.Line()
		instructionDiscriminatorsCodes.Var().Parens(
			DoGroup(
				func(code *Group) {
					for _, instruction := range g.idl.Instructions {
						if instruction.Discriminator == nil {
							continue
						}

						discriminator := instruction.Discriminator
						if len(discriminator) != 8 {
							panic(fmt.Errorf("discriminator for instruction %s must be exactly 8 bytes long", instruction.Name))
						}

						discriminatorName := FormatInstructionDiscriminatorName(instruction.Name)
						{
							code.Id(discriminatorName).Op("=").Index(Lit(8)).Byte().Op("{").ListFunc(func(byteGroup *Group) {
								for _, byteVal := range discriminator[:] {
									byteGroup.Lit(int(byteVal))
								}
							}).Op("}")
						}
						code.Line()
					}
				},
			),
		)
		file.Add(instructionDiscriminatorsCodes)
		file.Line()
	}
	return &OutputFile{
		Name: "discriminators.go",
		File: file,
	}, nil
}
