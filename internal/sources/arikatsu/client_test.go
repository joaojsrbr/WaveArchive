package arikatsu

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexesNormalizeRawArikatsuData(t *testing.T) {
	root := t.TempDir()
	versionRoot := filepath.Join(root, "3.5")
	writeFixture(t, versionRoot, "Textmaps/pt/multi_text/MultiText.json", `[
		{"Id":"Role_Name","Content":"Ressonador"},
		{"Id":"Role_Nick","Content":"Apelido"},
		{"Id":"Weapon_Name","Content":"Espada"},
		{"Id":"Weapon_Desc","Content":"Descrição"},
		{"Id":"Weapon_Story","Content":"História da arma"},
		{"Id":"Weapon_Passive","Content":"Passiva oficial"},
		{"Id":"Property_ATK","Content":"ATQ"},
		{"Id":"Echo_Name","Content":"Eco"},
		{"Id":"MonsterInfo_350000010_Name","Content":"Fantasma: Eco"},
		{"Id":"Sonata_Name","Content":"Sonata"},
		{"Id":"Sonata_2","Content":"Efeito de duas peças"}
	]`)
	writeFixture(t, versionRoot, "BinData/role/roleinfo.json", `[
		{"Id":1102,"QualityId":5,"RoleType":1,"Name":"Role_Name","NickName":"Role_Nick","ElementId":1,"WeaponType":2,"RoleBody":"FemaleM","RoleHeadIconLarge":"/Game/Aki/UI/a.a","FormationRoleCard":"/Game/Aki/UI/b.b","ShowInBag":true},
		{"Id":9999,"QualityId":3,"RoleType":2,"Name":"Hidden","ShowInBag":false}
	]`)
	writeFixture(t, versionRoot, "BinData/weapon/weaponconf.json", `[
		{"ItemId":21020001,"IsShow":true,"WeaponName":"Weapon_Name","QualityId":5,"WeaponType":2,"FirstPropId":{"Id":7,"Value":40},"FirstCurve":1,"SecondPropId":{"Id":10007,"Value":0.04,"IsRatio":true},"SecondCurve":2,"ResonId":21020001,"Desc":"Weapon_Desc","AttributesDescription":"Weapon_Story","Icon":"/Game/Aki/UI/w.w","ShowInBag":true}
	]`)
	writeFixture(t, versionRoot, "BinData/property/weaponpropertygrowth.json", `[
		{"CurveId":1,"Level":1,"BreachLevel":0,"CurveValue":10000},
		{"CurveId":1,"Level":90,"BreachLevel":6,"CurveValue":125000},
		{"CurveId":2,"Level":1,"BreachLevel":0,"CurveValue":10000},
		{"CurveId":2,"Level":90,"BreachLevel":6,"CurveValue":45000}
	]`)
	writeFixture(t, versionRoot, "BinData/property/propertyindex.json", `[
		{"Id":7,"Key":"Atk","Name":"Property_ATK"},
		{"Id":10007,"Key":"GreenAtk","Name":"Property_ATK","ConvertToWhiteId":7}
	]`)
	writeFixture(t, versionRoot, "BinData/weapon/weaponreson.json", `[
		{"ResonId":21020001,"Level":1,"Name":"Weapon_Passive"}
	]`)
	writeFixture(t, versionRoot, "BinData/phantom/phantomitem.json", `[
		{"ItemId":60000012,"MonsterId":390000001,"MonsterName":"Echo_Name","SkillId":200001,"Rarity":1,"QualityId":2,"Icon":"/Game/Aki/UI/e.e","FetterGroup":[7],"PhantomType":1},
		{"ItemId":60000015,"MonsterId":390000001,"MonsterName":"Echo_Name","SkillId":200001,"Rarity":1,"QualityId":5,"Icon":"/Game/Aki/UI/e.e","FetterGroup":[7],"PhantomType":1},
		{"ItemId":60000025,"MonsterId":390000002,"MonsterName":"MonsterInfo_350000010_Name","SkillId":200002,"Rarity":1,"QualityId":5,"Icon":"/Game/Aki/UI/cosmetic.e","FetterGroup":[7],"PhantomType":1}
	]`)
	writeFixture(t, versionRoot, "BinData/phantom/phantomfettergroup.json", `[
		{"Id":7,"FetterGroupName":"Sonata_Name","FetterElementPath":"/Game/Aki/UI/s.s","FetterMap":[{"Key":2,"Value":72}]}
	]`)
	writeFixture(t, versionRoot, "BinData/phantom/phantomfetter.json", `[
		{"Id":72,"EffectDescription":"Sonata_2","EffectDescriptionParam":[]}
	]`)

	client, err := NewClient("3.5", root, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	characters, err := client.CharacterIndex(ctx, "3.5")
	if err != nil || len(characters) != 1 || characters["1102"].Name != "Ressonador" || characters["1102"].Gender != "Female" {
		t.Fatalf("characters = %#v, %v", characters, err)
	}
	weapons, err := client.WeaponIndex(ctx, "3.5")
	if err != nil || len(weapons) != 1 || weapons["21020001"].Name != "Espada" ||
		weapons["21020001"].BaseATK != 500 || weapons["21020001"].SubStat != "ATQ 18%" ||
		weapons["21020001"].Description != "História da arma" {
		t.Fatalf("weapons = %#v, %v", weapons, err)
	}
	weapon, err := client.Weapon(ctx, "3.5", "pt", 21020001)
	if err != nil || weapon.Description != "História da arma" || weapon.EffectName != "Passiva oficial" || weapon.Effect != "Descrição" {
		t.Fatalf("weapon detail = %#v, %v", weapon, err)
	}
	echoes, err := client.EchoIndex(ctx, "3.5")
	if err != nil || len(echoes) != 1 || echoes["390000001"].Name != "Eco" || len(echoes["390000001"].Ranks) != 2 {
		t.Fatalf("echoes = %#v, %v", echoes, err)
	}
	sonatas, err := client.SonataIndex(ctx, "3.5")
	if err != nil || len(sonatas) != 1 || sonatas["7"].Name.English != "Sonata" {
		t.Fatalf("sonatas = %#v, %v", sonatas, err)
	}
}

func TestRejectsUnsupportedVersion(t *testing.T) {
	if _, err := NewClient("3.2", t.TempDir(), nil, nil); err == nil {
		t.Fatal("expected unsupported version error")
	}
}

func TestCachedRealBranchIndexes(t *testing.T) {
	root := os.Getenv("WAVEARCHIVE_ARIKATSU_TEST_ROOT")
	if root == "" {
		t.Skip("set WAVEARCHIVE_ARIKATSU_TEST_ROOT to a cached public branch")
	}
	client, err := NewClient("3.5", root, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	characters, err := client.CharacterIndex(ctx, "3.5")
	if err != nil || len(characters) < 50 {
		t.Fatalf("real character index = %d, %v", len(characters), err)
	}
	weapons, err := client.WeaponIndex(ctx, "3.5")
	if err != nil || len(weapons) < 50 {
		t.Fatalf("real weapon index = %d, %v", len(weapons), err)
	}
	if weapon := weapons["21050084"]; weapon.BaseATK != 463 || weapon.SubStat != "ATQ 18,2%" {
		t.Fatalf("Nuclear Accretion stats = %#v", weapon)
	}
	echoes, err := client.EchoIndex(ctx, "3.5")
	if err != nil || len(echoes) != 180 {
		t.Fatalf("real Echo index = %d, %v", len(echoes), err)
	}
	sonatas, err := client.SonataIndex(ctx, "3.5")
	if err != nil || len(sonatas) < 10 {
		t.Fatalf("real Sonata index = %d, %v", len(sonatas), err)
	}
}

func writeFixture(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
