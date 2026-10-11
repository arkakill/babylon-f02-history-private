# Preflight técnico — matriz histórica F-02

Fecha: 2026-10-10.

## Copia de trabajo

- Repositorio local independiente: `work/private-audit-copy`.
- Rama: `audit/f02-history-reconstruction`.
- HEAD antes del nuevo commit: `961c6a9d09ad6949441904b0c9ac81b0fa8a4882`.
- Padre: `02a6083bc7a7a4b2223ee93914edea56bc04b98f`.
- El cambio tracked en el árbol de trabajo se limita a `.github/workflows/f02-history-matrix.yml`; los informes y resultados previos se agregan bajo `audit-evidence/`.
- No se modificó código de producción, `go.mod` ni `go.sum`. No se ejecutaron pruebas Go localmente.

## Verificaciones locales realizadas

1. `git -c "safe.directory=<private-audit-copy>" -C <private-audit-copy> diff --cached --check -- .github/workflows/f02-history-matrix.yml audit-evidence/*.md audit-evidence/*.json` — código 0, sin diagnósticos en el workflow ni en los informes/JSON.
2. El `git diff --cached --check` global devuelve código 1 por espacios y tabulaciones presentes en los logs de pruebas archivados. Se conservan byte por byte porque son salida original de GitHub Actions; no se editaron para limpiar el formato.
3. `actionlint.exe -version` — `1.7.12`, Windows/amd64, Go 1.26.1.
4. `actionlint.exe <private-audit-copy>/.github/workflows/f02-history-matrix.yml` — código 0, sin diagnósticos. El ZIP oficial de Windows x64 fue validado contra el archivo de checksums oficial; SHA-256 `6e7241b51e6817ea6a047693d8e6fed13b31819c9a0dd6c5a726e1592d22f6e9`.

## Pendiente

La autenticación de GitHub sigue bloqueada en el entorno local. `gh` no se encuentra en PATH, por lo que `gh auth status` no se pudo ejecutar en esta comprobación. `git credential-manager github list` termina con código 1: “Unable to persist credentials with the 'wincredman' credential store.” La pestaña de activación está en `github.com/login/device`, iniciada como `arkakill`, y presenta el formulario de código vacío; esto no acredita que la CLI haya quedado autorizada.

Hasta resolver ese acceso no se puede verificar/crear el repositorio privado, publicar esta rama, ni lanzar o descargar los resultados nuevos de GitHub Actions. No se han simulado esas operaciones. No se ejecutaron pruebas Go en Windows.
