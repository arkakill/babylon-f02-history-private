# Estado de la reconstrucción histórica F-02 — 2026-10-10

## Veredicto histórico

**NO RESUELTO HISTÓRICAMENTE.** Los commits candidatos de PR #631 y PR #634 están identificados, pero la matriz corregida no se pudo publicar ni ejecutar. El intento anterior de la matriz no llegó a las aserciones finales en esos cuatro commits por incompatibilidades del harness. Por tanto, no hay base de prueba para llamar a la corrección verificada, incompleta o regresión.

Esto no revoca la reproducción ya confirmada de F-02 en `v4.5.0` y en el SHA de upstream `main` examinado. En ambos, el estado derivado terminó `EXPIRED` mientras la tabla de poder finality conservó `2.000.000.000` sats tras el quórum tardío.

## Commits de la matriz

| Revisión | SHA exacto | Control | Test F-02 | Estado y poder final | Evidencia / conclusión |
|---|---|---|---|---|---|
| Padre PR #631 | `cbc3a9a8c0c1e143f44f819dc145051dabf0f122` | No verificado: intento anterior bloqueado antes de invariantes | No verificado: compilación/adaptación del harness no llegó a la aserción | Desconocidos | PR #631; falta matriz corregida |
| Corrección PR #631 | `6d9b66816d33bd5bbb2abc9a55eb29324d836715` | No verificado: intento anterior bloqueado antes de invariantes | No verificado: compilación/adaptación del harness no llegó a la aserción | Desconocidos | Commit de PR #631; falta matriz corregida |
| Padre PR #634 | `b68ef264564cfb01dd0d270076cec4055f3d1c6d` | No verificado: intento anterior bloqueado antes de invariantes | No verificado: compilación/adaptación del harness no llegó a la aserción | Desconocidos | PR #634; falta matriz corregida |
| Backport PR #634 | `1ebc3727fba1bf6231cb11d70554b12fa58d2342` | No verificado: intento anterior bloqueado antes de invariantes | No verificado: compilación/adaptación del harness no llegó a la aserción | Desconocidos | Commit backport de PR #634; falta matriz corregida |
| Babylon v4.5.0 | `f9006e4232dece8050b08556c4617bc0aaa07128` | PASS; antes de expirar, estado `ACTIVE`, cache y tabla con `2.000.000.000` sats | F-02 reproducida; código 1 porque falla la aserción de poder cero | Tras expiración normal: `EXPIRED`, poder 0. En la secuencia objetivo: evento `EXPIRED` consumido sin quórum (`PENDING`, tabla vacía); tras firmas tardías y evento `ACTIVE`: `EXPIRED`, tabla con `2.000.000.000` sats | Ejecución previa de Actions archivada; [run 38095710474](https://github.com/arkakill/babylon/actions/runs/38095710474). Reproducción confirmada en este SHA |
| Upstream main examinado el 2026-10-10 | `c2f731f4531b2b22afe5047f152e108b944bf811` | PASS; antes de expirar, estado `ACTIVE`, cache y tabla con `2.000.000.000` sats | F-02 reproducida; código 1 porque falla la aserción de poder cero | Tras expiración normal: `EXPIRED`, poder 0. En la secuencia objetivo: evento `EXPIRED` consumido sin quórum (`PENDING`, tabla vacía); tras firmas tardías y evento `ACTIVE`: `EXPIRED`, cache y tabla con `2.000.000.000` sats | [Run 38098547912](https://github.com/arkakill/babylon/actions/runs/38098547912), artefacto y logs preservados. Reproducción confirmada en este SHA |

Los dos estados “no verificado” de cada commit candidato se refieren al intento fallido de matriz [38100558671](https://github.com/arkakill/babylon/actions/runs/38100558671), no a un resultado de seguridad. El resumen global `Success` de ese run no convierte en prueba los trabajos que registraron `NOT_VERIFIED`.

## Reconstrucción del cambio candidato

- PR #631 se titula “ignore expire events when there is no covenant quorum”. Su cambio evita procesar la resta de poder al vencer una delegación si aún no tiene quórum. El test que acompaña al cambio cubre expiración sin quórum, no la llegada de quórum después de consumir `EXPIRED`.
- PR #634 identifica su cambio como backport de #631 a `release/v1.x`.
- Los SHA de los commits y sus padres se verificaron contra la API/historial oficial y se conservan en [historical-candidates.md](historical-candidates.md).
- No se halló una atribución pública explícita que conecte esos dos PR con la remediación del hallazgo de OpenZeppelin. La similitud temática no demuestra esa atribución.

Referencias: [PR #631](https://github.com/babylonlabs-io/babylon/pull/631), [PR #634](https://github.com/babylonlabs-io/babylon/pull/634), [OpenZeppelin](https://www.openzeppelin.com/news/state-changes-at-the-boundary-lessons-from-security-research-on-babylon).

## Pruebas y comandos

Comandos definidos en el workflow, pendientes de ejecución nueva en Linux:

```text
go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -count=1 -v -run '^TestF02ControlQuorumBeforeExpiry$' ./x/finality/keeper
go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -count=1 -v -run '^TestF02LateQuorumAfterExpiryHasNoVotingPower$' ./x/finality/keeper
```

Comprobaciones locales realizadas en Windows:

- `actionlint.exe <private-audit-copy>/.github/workflows/f02-history-matrix.yml` — código de salida `0`.
- `git diff --cached --check -- .github/workflows/f02-history-matrix.yml audit-evidence/*.md audit-evidence/*.json` — código `0`.
- `git diff --cached --check` global — código `1` únicamente por whitespace que aparece en logs originales de Actions preservados como evidencia.
- No se ejecutaron pruebas Go locales ni compilaciones.

## Bloqueo y siguiente paso

La CLI `gh` no está disponible en PATH. Git Credential Manager 2.7.3 está instalado, pero su consulta de cuentas acabó con código `1` porque no puede persistir credenciales en `wincredman`. En el navegador aparece la pantalla de activación de dispositivo con el usuario `arkakill` y el campo de código vacío. No se verificó una sesión autenticada de CLI, no se creó ni verificó un repositorio privado y no se hizo push. Tampoco se lanzó un workflow nuevo.

Siguiente paso: restablecer autenticación local autorizada de GitHub (instalar/disponer de `gh` o habilitar Credential Manager en un contexto donde pueda guardar la sesión), crear/verificar el destino privado `arkakill/babylon-f02-history-private`, empujar únicamente `audit/f02-history-reconstruction`, y descargar los artefactos completos de Actions. Después se podrá clasificar la historia comparando padre y corrección con la misma prueba.

## Alcance y certeza

- **F-02 en v4.5.0 y `c2f731f`: certeza alta**, respaldada por pruebas de integración en Actions y almacenamiento de keeper observado.
- **Efecto de PR #631/#634 sobre esta secuencia: inconcluso**, porque las pruebas no llegaron a las aserciones en esos commits.
- **Regresión frente a una corrección de OpenZeppelin: no demostrada.** Tampoco se atribuye públicamente la remediación a PR #631/#634.
- No se usaron redes Bitcoin, transacciones, claves, fondos, secretos ni cambios a producción o dependencias.
