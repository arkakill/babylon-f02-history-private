# Auditoría comparativa F-02: Babylon BTC Staking

**Veredicto: DUPLICADA/REGRESIÓN — duplicada del hallazgo público de OpenZeppelin y confirmada en `c2f731f4531b2b22afe5047f152e108b944bf811`; una regresión histórica respecto de una corrección concreta no está demostrada.**

## Resultado en palabras simples

Probé la misma secuencia en el commit exacto de `main` indicado abajo. El test de control pasó. En el test objetivo, Babylon procesó la expiración cuando todavía faltaban firmas covenant, consumió ese evento y después aceptó el quórum. En el bloque siguiente, la delegación seguía derivándose como `EXPIRED`, pero la tabla de poder de voto guardó **2.000.000.000 sats** para un proveedor.

La prueba no se limita a calcular un resultado en un modelo: usa los keepers, el `MsgServer`, el `BeginBlocker`, el almacén de eventos y la tabla/cache KV del harness de integración del propio repositorio. Los servicios externos que aportan datos de Bitcoin y checkpoint sí están simulados. No se enviaron transacciones a redes.

El comportamiento coincide sustancialmente con el hallazgo publicado por OpenZeppelin en abril de 2026. No encontré una referencia pública que asocie ese hallazgo a un commit correctivo exacto de Babylon. Por eso puedo confirmar que la anomalía está en el commit examinado y que el hallazgo es duplicado, pero no afirmar que sea una regresión tras una corrección conocida.

## Commit, rama y cambios

- Repositorio oficial: [babylonlabs-io/babylon](https://github.com/babylonlabs-io/babylon).
- Fork de laboratorio: [arkakill/babylon](https://github.com/arkakill/babylon).
- Snapshot exacto de `main` examinado: [`c2f731f4531b2b22afe5047f152e108b944bf811`](https://github.com/babylonlabs-io/babylon/commit/c2f731f4531b2b22afe5047f152e108b944bf811), fechado el 2 de octubre de 2026. Es un SHA fijo, no una referencia móvil.
- Rama independiente del fork: `audit/f02-upstream-main-c2f731f`.
- Commit de auditoría probado por el workflow final: [`15aad2ae9c2aec1f3f1f1331200fab01a8043131`](https://github.com/arkakill/babylon/commit/15aad2ae9c2aec1f3f1f1331200fab01a8043131).
- El artifact final registra como base común exactamente `c2f731f4531b2b22afe5047f152e108b944bf811` y lista solo estos archivos: el workflow de auditoría y el test F-02.
- El archivo del test conserva exactamente los bytes usados en v4.5.0 (SHA-256 `6b961b5003cc9d348dd8e71aebcdb9829f6869cadcee670399fc0fb388bf6116`); no hubo que adaptar helpers.
- No se modificaron código de producción, `go.mod` ni `go.sum`.

## Hipótesis y causa observada

F-02 plantea que una delegación Bitcoin expirada puede volver a recibir poder de voto si se alcanza el quórum covenant después de que el evento `EXPIRED` ya fue consumido.

La ruta observada es:

1. Sin quórum, el estado derivado queda `PENDING` antes de que se compruebe el vencimiento.
2. El `BeginBlocker` procesa y elimina el evento de expiración sin restar poder que nunca se había contado.
3. Más tarde `MsgServer.AddCovenantSigs` completa el quórum y crea un evento `ACTIVE` en el tip Bitcoin actual, aunque la delegación ya se deriva como `EXPIRED`.
4. El procesador de distribución aplica ese `ACTIVE` sin volver a validar el estado derivado. El evento de expiración anterior ya no está para quitar el poder.

El código examinado conserva la comprobación de quórum antes de la lógica de vencimiento en [`BTCDelegation.GetStatus`](https://github.com/babylonlabs-io/babylon/blob/c2f731f4531b2b22afe5047f152e108b944bf811/x/btcstaking/types/btc_delegation.go). La finalización del quórum añade el evento activo en [`btc_delegations.go`](https://github.com/babylonlabs-io/babylon/blob/c2f731f4531b2b22afe5047f152e108b944bf811/x/btcstaking/keeper/btc_delegations.go), y la rama activa del procesador de poder lo aplica en [`power_dist_change.go`](https://github.com/babylonlabs-io/babylon/blob/c2f731f4531b2b22afe5047f152e108b944bf811/x/finality/keeper/power_dist_change.go).

## Pruebas ejecutadas

GitHub Actions ejecutó las dos pruebas por separado en `ubuntu-22.04`, con Go `1.25.8` (`linux/amd64`) y los tags del repositorio. El workflow usa `-mod=readonly`.

```bash
go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -count=1 -v -run '^TestF02ControlQuorumBeforeExpiry$' ./x/finality/keeper
go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -count=1 -v -run '^TestF02LateQuorumAfterExpiryHasNoVotingPower$' ./x/finality/keeper
```

| Prueba | Resultado Go | Código de salida | Interpretación |
|---|---:|---:|---|
| `TestF02ControlQuorumBeforeExpiry` | PASS | 0 | Quórum previo: la delegación está `ACTIVE` con 2.000.000.000 sats; al expirar, poder 0 y tabla vacía. |
| `TestF02LateQuorumAfterExpiryHasNoVotingPower` | FAIL en la aserción de seguridad | 1 | El test compiló y recorrió el escenario; observó 2.000.000.000 sats donde esperaba cero. Esta falla es el resultado que buscaba detectar el test. |

### Valores observados en la prueba objetivo

| Bloque Babylon / altura BTC | Evento y estado derivado | Poder guardado |
|---|---|---:|
| 4 / 810 | Se procesa `EXPIRED` sin quórum; el evento queda consumido, pero `GetStatus` deriva `PENDING`. Cache vacío y tabla `map[]`. | 0 |
| 5 / 811 | El quórum se completa mediante el `MsgServer`; el estado derivado ya es `EXPIRED` y se emite un evento `ACTIVE`. | 0 antes de procesar ese evento |
| 6 / 811 | El `BeginBlocker` aplica `ACTIVE`; el estado derivado sigue `EXPIRED`. Cache y tabla contienen el poder. | **2.000.000.000 sats** |

Estado final leído del almacenamiento KV del keeper:

- Cache: `total_voting_power=2000000000`, `num_active_fps=1`, una entrada de proveedor con `total_bonded_sat=2000000000` y `is_timestamped=true`.
- Tabla completa: `8b928a1aec3f967c684b26487784c1b87e216665615d0eea7058186414874ed8 → 2000000000`.
- La salida registra `derived_state=EXPIRED` al mismo tiempo que el poder está en 2.000.000.000.

## Runs y archivos de evidencia

### Ejecución final con la base exacta

[GitHub Actions run 38098547912](https://github.com/arkakill/babylon/actions/runs/38098547912), commit de auditoría `15aad2ae9c2aec1f3f1f1331200fab01a8043131`. El resumen del job registra:

- upstream base SHA: `c2f731f4531b2b22afe5047f152e108b944bf811`;
- control: código 0;
- objetivo: código 1;
- artefacto `f02-main-c2f731f-audit-results`, retenido en Actions 90 días;
- SHA-256 del ZIP, verificado localmente y coincidente con GitHub: `5c591b088ee86944a95e34c47d49e4fdb5f086350c1e2631bdbe61c67ac2f71d`.

Descarga local conservada: [ZIP con resultados completos](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912.zip>). Logs y metadatos extraídos: [control.log](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912/control.log>), [target.log](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912/target.log>), [control.exit-code](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912/control.exit-code>), [target.exit-code](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912/target.exit-code>), [upstream-base.sha](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912/upstream-base.sha>) y [audit-commit-files.txt](</C:/Users/eliga/Documents/Codex/2026-10-10/como/outputs/f02-main-c2f731f-run-38098547912/audit-commit-files.txt>).

### Ejecución inicial conservada

[GitHub Actions run 38098321795](https://github.com/arkakill/babylon/actions/runs/38098321795), commit `8c4bbed8fd2bf282af6abd7dd3b62b5aa18794f7`. Ambas pruebas también se ejecutaron y dieron 0/1, pero el checkout superficial dejó vacío el campo de base SHA. Conservé su artifact para el historial; no lo uso como evidencia de la base exacta. Corregí el workflow para descargar el historial completo y comprobar que la base común fuera exactamente `c2f731f…`; el run final volvió a ejecutar ambas pruebas.

## Comparación con antecedentes públicos

- **OpenZeppelin, 1 de abril de 2026:** su artículo describe explícitamente la secuencia `EXPIRED → ACTIVE`, poder de voto persistente y el riesgo para las premisas de seguridad de finality. También afirma que Babylon confirmó los hallazgos y publicó correcciones. La coincidencia de secuencia, causa e impacto hace que F-02 sea una duplicación sustancial de ese hallazgo. El artículo no identifica públicamente el commit/release correctivo específico: [State Changes at the Boundary](https://www.openzeppelin.com/news/state-changes-at-the-boundary-lessons-from-security-research-on-babylon).
- **Babylon PR #631, 5 de marzo de 2025:** `ignore expire events when there is no covenant quorum` (commit principal enlazado desde el PR: [`6d9b668`](https://github.com/babylonlabs-io/babylon/commit/6d9b668)). Atiende el caso de una expiración cuando el stake nunca alcanzó quórum. No cubre la secuencia probada aquí: expiración consumida y quórum conseguido más tarde, que origina un nuevo `ACTIVE`: [PR #631](https://github.com/babylonlabs-io/babylon/pull/631).
- **Babylon PR #1404, 25 de julio de 2025:** refactorizó el procesamiento de distribución de poder y agregó una prueba de `ACTIVE` y `EXPIRED` en el mismo bloque. Su descripción lo presenta como una refactorización de determinismo/mantenibilidad; el diff no identifica una corrección de F-02. El escenario de esta auditoría cruza bloques: el evento `EXPIRED` se consume antes de que aparezca el evento `ACTIVE`: [PR #1404](https://github.com/babylonlabs-io/babylon/pull/1404).
- **Comparación v4.5.0:** el mismo test byte por byte reprodujo antes el mismo resultado en `f9006e4232dece8050b08556c4617bc0aaa07128` (run [38095710474](https://github.com/arkakill/babylon/actions/runs/38095710474)). La presencia en ambos extremos confirma que no es una anomalía nueva de `main`. No descarta que una corrección no localizada se haya revertido en algún punto intermedio; no hay evidencia suficiente para afirmar esa historia.

## Limitaciones y certeza

- **Alta certeza** de reproducción en el commit `c2f731f…`: el control pasó, la prueba objetivo compiló y falló solo porque encontró poder distinto de cero; el estado derivado, cache, tabla y código de salida están en el artifact.
- **Alta certeza** de que el hallazgo es duplicado sustancial del reporte público de OpenZeppelin.
- **No resuelto** si la corrección que Babylon asocia a OpenZeppelin fue incompleta, se aplicó en una rama/version distinta o se revirtió: no se localizó una referencia pública exacta a ese patch.
- El test usa el harness real de keepers y almacenes KV, pero no ejecuta un nodo CometBFT, un commit IAVL de disco ni una red Bitcoin. Los keepers de light client/checkpoint son mocks de fixture.
- Las pruebas se ejecutaron en GitHub Actions, no en el Windows de 8 GB. No se hicieron compilaciones ni pruebas masivas locales.
- El workflow final advierte que ciertas actions antiguas usan Node 20 y GitHub las fuerza a Node 24; el job llegó a ejecutar las pruebas y guardar el artifact, así que la advertencia no impidió este resultado.

## Recomendación automática siguiente

Identificar mediante referencias verificables el commit o release exacto que Babylon atribuye a F-02/OpenZeppelin. Cuando se encuentre, ejecutar este mismo test sin modificarlo en ese commit, el padre y el primer release posterior; si una versión intermedia pasa y otra posterior vuelve a fallar, comparar el diff que cambió `GetStatus`, la emisión de `ACTIVE` y el procesamiento de la tabla. Hasta entonces, registrar el estado como **duplicado confirmado en `main`, regresión histórica no demostrada**.
