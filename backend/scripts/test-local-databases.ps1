param(
    [string]$MongoBin = 'C:\Program Files\MongoDB\Server\9.0\bin',
    [string]$PostgresBin = 'C:\Program Files\PostgreSQL\18\bin',
    [switch]$Browser
)
$ErrorActionPreference = 'Stop'
# One owner creates BOTH new clusters, isolates the environment and restores it.
# Migration ownership remains in test-local-postgres -> official Go migrator.
& (Join-Path $PSScriptRoot 'test-local-postgres.ps1') -PostgresBin $PostgresBin -MongoBin $MongoBin -Browser:$Browser
