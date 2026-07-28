package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"

	"wavearchive/internal/assets"
	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
	"wavearchive/internal/sources/arikatsu"
	"wavearchive/internal/sources/nanoka"
	"wavearchive/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("wavearchive-cli", flag.ContinueOnError)
	dbPath := flags.String("db", filepath.Join(configDir, "WaveArchive", "wavearchive.db"), "caminho do banco SQLite")
	sourceName := flags.String("source", "nanoka", "fonte do catálogo: nanoka ou arikatsu")
	version := flags.String("version", "", "snapshot exato; obrigatório para Arikatsu")
	noAssets := flags.Bool("no-assets", false, "não baixar imagens durante sync")
	if err := flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	args := flags.Args()
	if len(args) == 0 {
		printUsage()
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		return err
	}

	resolvedDBPath, err := database.ResolveApplicationPath(*dbPath)
	if err != nil {
		return err
	}
	db, err := database.Open(resolvedDBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	repo := repository.NewCharacterSQLite(db.SQL())
	weaponRepo := repository.NewWeaponSQLite(db.SQL())
	echoRepo := repository.NewEchoSQLite(db.SQL())
	var source usecase.CatalogSource = nanoka.NewClient(nil)
	if *sourceName == "arikatsu" {
		if *version == "" {
			*version = "3.5"
		}
		source, err = arikatsu.NewClient(*version, filepath.Join(filepath.Dir(resolvedDBPath), "sources", "arikatsu"), nil, nanoka.NewClient(nil))
		if err != nil {
			return err
		}
	} else if *sourceName != "nanoka" {
		return fmt.Errorf("fonte desconhecida %q", *sourceName)
	}
	var assetCache *assets.Cache
	if !*noAssets {
		assetCache = assets.NewCache(filepath.Join(filepath.Dir(resolvedDBPath), "assets"), nil)
	}
	catalog := usecase.NewCharacterCatalog(repo, source, assetCache, logger)
	weaponCatalog := usecase.NewWeaponCatalog(weaponRepo, source, assetCache, logger)
	echoCatalog := usecase.NewEchoCatalog(echoRepo, source, assetCache, logger)
	ctx := context.Background()

	switch args[0] {
	case "sync":
		syncProgress := func(stage string, progress int) {
			fmt.Fprintf(os.Stderr, "\r%-14s %3d%%", stage, progress)
		}
		var result domain.SyncResult
		if *version == "" {
			result, err = catalog.Sync(ctx, syncProgress)
		} else {
			result, err = catalog.SyncVersion(ctx, *version, syncProgress)
		}
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		weaponCount, err := weaponCatalog.Sync(ctx, result.Version, func(stage string, progress int) {
			fmt.Fprintf(os.Stderr, "\r%-18s %3d%%", stage, progress)
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		echoCount, err := echoCatalog.Sync(ctx, result.Version, func(stage string, progress int) {
			fmt.Fprintf(os.Stderr, "\r%-18s %3d%%", stage, progress)
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return err
		}
		fmt.Printf("Sincronizados %d personagens, %d armas e %d Echoes da versão %s.\n", result.Count, weaponCount, echoCount, result.Version)
	case "list":
		return printCharacters(ctx, catalog)
	case "show":
		if len(args) < 2 {
			return errors.New("informe o ID: wavearchive-cli show <id>")
		}
		id, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("ID inválido: %w", err)
		}
		profile, err := catalog.Get(ctx, id)
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "list-weapons":
		weapons, err := weaponCatalog.List(ctx, domain.WeaponFilter{Sort: "id"})
		if err != nil {
			return err
		}
		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(writer, "ID\tNOME\tTIPO\tRARIDADE\tATK\tSUBSTAT")
		for _, weapon := range weapons {
			fmt.Fprintf(writer, "%d\t%s\t%s\t%d★\t%d\t%s\n", weapon.ID, weapon.Name, weapon.TypeName, weapon.Rarity, weapon.BaseATK, weapon.SubStat)
		}
		return writer.Flush()
	case "list-echoes":
		echoes, err := echoCatalog.List(ctx, domain.EchoFilter{Sort: "id"})
		if err != nil {
			return err
		}
		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(writer, "ID\tNOME\tCLASSE\tCUSTO\tPOSSUÍDOS")
		for _, echo := range echoes {
			fmt.Fprintf(writer, "%d\t%s\t%s\t%d\t%d\n", echo.ID, echo.Name, echo.Class, echo.Cost, echo.OwnedCount)
		}
		return writer.Flush()
	default:
		return fmt.Errorf("comando desconhecido %q", args[0])
	}
	return nil
}

func printCharacters(ctx context.Context, catalog *usecase.CharacterCatalog) error {
	characters, err := catalog.List(ctx, domain.CharacterFilter{Sort: "id"})
	if err != nil {
		return err
	}
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tNOME\tELEMENTO\tRARIDADE\tARMA")
	for _, character := range characters {
		fmt.Fprintf(writer, "%d\t%s\t%s\t%d★\t%s\n", character.ID, character.Name, character.ElementName, character.Rarity, character.WeaponTypeName)
	}
	return writer.Flush()
}

func printUsage() {
	fmt.Println(`WaveArchive CLI

Uso:
  wavearchive-cli [-db caminho] [-source nanoka|arikatsu] [-version 3.5] sync
  wavearchive-cli [-db caminho] list
  wavearchive-cli [-db caminho] list-weapons
  wavearchive-cli [-db caminho] list-echoes
  wavearchive-cli [-db caminho] show <id>`)
}
