import {describe,expect,it} from "vitest";
import {parseArchive,validateArchiveHeader} from "./archive";

describe("archive validation",()=>{
  it("accepts schema 1",()=>expect(validateArchiveHeader({schemaVersion:1,exportedAt:"2026-07-26"}).schemaVersion).toBe(1));
  it("rejects unknown schemas",()=>expect(()=>parseArchive('{"schemaVersion":9,"exportedAt":"x"}')).toThrow(/não suportada/));
  it("rejects empty input",()=>expect(()=>parseArchive(" ")).toThrow(/vazio/));
});
