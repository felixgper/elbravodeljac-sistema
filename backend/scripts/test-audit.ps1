param([switch]$SkipContainerSuite,[switch]$StartIsolated,[string]$TestProject,[string]$TestOverride,[int]$PostgresPort,[int]$MongoPort)
$ErrorActionPreference='Stop'
# Keep the historical entry point, but never migrate/reuse a pre-existing QA.
if ($PSBoundParameters.Count) {
    throw 'Los parámetros históricos de reutilización están retirados. Usa test-isolated.ps1 -Browser; no se modificó ningún entorno.'
}
& (Join-Path $PSScriptRoot 'test-isolated.ps1') -Browser
