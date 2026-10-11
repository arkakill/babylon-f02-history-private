# Auditoría técnica de Babylon BTC Staking: F-02

**Veredicto: DUPLICADA/REGRESIÓN — hallazgo duplicado reproducido; regresión histórica no demostrada.**

## Resumen ejecutivo

En la etiqueta v4.5.0 y el commit base esperado f9006e4232dece8050b08556c4617bc0aaa07128, la secuencia investigada sí se reproduce en GitHub Actions Linux. La prueba de control —quórum antes de vencer— pasa. En la prueba objetivo se procesa el evento de expiración sin quórum, después se completa el quórum y el sistema añade **2,000,000,000 sats de poder de voto persistido**, aunque el estado derivado de la delegación indica EXPIRED.

El resultado coincide sustancialmente con un hallazgo público de OpenZeppelin publicado el 1 de abril de 2026. El artículo declara que Babylon confirmó los hallazgos y publicó correcciones, pero la evidencia consultada no identifica el commit o release exacto que corrige este hallazgo. La reproducción está confirmada para v4.5.0; llamarla regresión respecto de una corrección específica no está probado.

## Identificación y alcance

- Fork de trabajo: [arkakill/babylon](https://github.com/arkakill/babylon); proyecto auditado: [babylonlabs-io/babylon](https://github.com/babylonlabs-io/babylon).
- Versión/tag base: v4.5.0; commit verificado: f9006e4232dece8050b08556c4617bc0aaa07128.
- Fecha de la etiqueta: 2026-09-29T08:59:53-03:00.
- Rama: audit/f-02-quorum-expiry-v4.5.0.
- Commit de auditoría final: df7d48889a989ebb1f8fe828bae3908dd8009ced.
- Cambios relativos a la etiqueta: solo el workflow y el test de auditoría. No se modificaron código de producción, go.mod ni go.sum.
- git diff --check pasó. El árbol quedó limpio después de los commits/push al fork.
- No se ejecutaron pruebas localmente en Windows. GitHub Actions ejecutó las pruebas en ubuntu-22.04 con Go 1.23.8.

## Hipótesis y causa técnica

F-02: una delegación Bitcoin recibe poder de voto después de expirar si el quórum covenant se completa tarde y el evento EXPIRED ya se consumió.

1. GetStatus prioriza la falta de quórum y devuelve PENDING antes de evaluar expiración.
2. AddCovenantSigs acepta la delegación PENDING; se pueden añadir firmas tardías.
3. Al completarse quórum, la ruta de delegación crea un evento ACTIVE en el tip Bitcoin actual sin rechazar explícitamente que venció.
4. El procesador de poder aplica ACTIVE sin comprobar de nuevo que el estado derivado sea ACTIVE. Suma el stake al cache y la tabla persistida, aunque GetStatus siga indicando EXPIRED. El evento EXPIRED anterior ya se consumió.

Referencias de v4.5.0: [GetStatus](https://github.com/babylonlabs-io/babylon/blob/v4.5.0/x/btcstaking/types/btc_delegation.go), [AddCovenantSigs](https://github.com/babylonlabs-io/babylon/blob/v4.5.0/x/btcstaking/keeper/msg_server.go), creación de eventos en [btc_delegations.go](https://github.com/babylonlabs-io/babylon/blob/v4.5.0/x/btcstaking/keeper/btc_delegations.go) y distribución en [power_dist_change.go](https://github.com/babylonlabs-io/babylon/blob/v4.5.0/x/finality/keeper/power_dist_change.go).

## Pruebas y resultados

Los tests recorren el test harness real de keepers, MsgServer, BeginBlocker, event store, cache y tabla de poder. Fixture: cinco firmantes covenant, quórum de tres y stake de 2,000,000,000 sats. Servicios externos Bitcoin/checkpoint simulados; el estado del keeper y almacenamiento de poder se verifican en el flujo de integración. No se usaron redes, transacciones, claves privadas ni fondos.

Comandos exactos:

~~~bash
go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -count=1 -v -run '^TestF02ControlQuorumBeforeExpiry$' ./x/finality/keeper
go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -count=1 -v -run '^TestF02LateQuorumAfterExpiryHasNoVotingPower$' ./x/finality/keeper
~~~

### Control: quórum antes de expirar

Resultado: **PASS**, código de salida 0.

- BTC height 809: estado derivado ACTIVE; poder persistido 2,000,000,000.
- BTC height 810: estado derivado EXPIRED; poder 0; tabla vacía.
- Cuando la distribución protobuf está vacía, el cache se lee ausente. El test comprueba poder cero y tabla vacía.

### Escenario objetivo: expiración sin quórum y quórum tardío

Resultado: **reproducido**; el test falla por el invariante de seguridad; código de salida 1.

| Punto observado | Estado derivado | Poder |
|---|---|---:|
| Babylon 4 / BTC 810: EXPIRED procesado sin quórum y evento consumido | PENDING | 0 |
| Babylon 5 / BTC 811: quórum completado por MsgServer | EXPIRED | 0 antes de aplicar evento |
| Babylon 6 / BTC 811: evento ACTIVE procesado | EXPIRED | **2,000,000,000** |

Estado persistido final:

- Cache: total_voting_power=2000000000, num_active_fps=1.
- Tabla completa: 8b928a1aec3f967c684b26487784c1b87e216665615d0eea7058186414874ed8:2000000000.
- La aserción esperaba cero, pero observó 2000000000.

El control pasó y la prueba objetivo compiló y ejecutó el flujo. La conclusión no se basa en sospecha estática ni en un test que no compiló.

## GitHub Actions y registros

Workflow añadido al fork: [f02-late-quorum-audit.yml](https://github.com/arkakill/babylon/blob/audit/f-02-quorum-expiry-v4.5.0/.github/workflows/f02-late-quorum-audit.yml). Ejecuta pruebas separadas en ubuntu-22.04 y conserva comandos, logs y códigos de salida en el artifact f02-audit-results por 90 días.

### Run final

[Actions run 38095710474](https://github.com/arkakill/babylon/actions/runs/38095710474), job 114341048978, commit df7d48889a989ebb1f8fe828bae3908dd8009ced.

- Control: exit 0.
- Objetivo: exit 1 por poder persistido positivo después de expiración.
- Artifact ID: 11686012464.
- SHA-256 del ZIP descargado: 5ce82cb9c8f90d4b74801d9a8a1332a8df62c57683b80eb96fb7fcb9abdd710d.
- El JSON contiene los logs completos y códigos de salida de ambas pruebas para las dos ejecuciones.

### Primera ejecución

[Actions run 38094833034](https://github.com/arkakill/babylon/actions/runs/38094833034), commit inicial de auditoría 5a40bfd54e0778b4335e6c13b70ca179c2939ec7.

No sirve para concluir sobre F-02: la aserción del harness exigía que el cache vacío fuera un objeto no nulo, aunque protobuf vacío se serializa sin bytes. Ambas pruebas se detenían en esa aserción. Se corrigió solo el test.

- Artifact ID: 11685976640.
- SHA-256 del ZIP descargado: f35087bdcf147528e046da53dc362c49168a8e9d52bb5dcf7cd83984b289584b.

## Comparación con antecedentes

- OpenZeppelin, [State Changes at the Boundary: Lessons from Security Research on Babylon](https://www.openzeppelin.com/news/state-changes-at-the-boundary-lessons-from-security-research-on-babylon), 2026-04-01: describe sustancialmente la misma prioridad de GetStatus, aceptación de firmas tardías, orden EXPIRED → ACTIVE y poder persistente. Dice que Babylon confirmó y corrigió hallazgos; justifica clasificar este caso como duplicado.
- Babylon [PR #631](https://github.com/babylonlabs-io/babylon/pull/631), merged 2025-03-05, relacionado con [commit 1ebc3727](https://github.com/babylonlabs-io/babylon/commit/1ebc3727): ignora eventos de expiración si aún no hay quórum. Esto cubre el evento EXPIRED prematuro, pero no el quórum completado después de consumirlo y la posterior aplicación de ACTIVE.
- Que v4.5.0 sea posterior al artículo no demuestra por sí solo regresión ni permite atribuir un parche ausente. No se concluye sobre otros tags ni sobre redes desplegadas.

## Limitaciones

- Solo se verificó v4.5.0; no se ejecutó una matriz de versiones ni se inspeccionaron redes desplegadas.
- Los servicios externos Bitcoin/checkpoint son simulados por el harness. Eventos, cache y tabla de poder del keeper se inspeccionaron en integración.
- Falta identificar el commit/release exacto atribuido por Babylon a la corrección pública; no se puede distinguir con certeza regresión de corrección incompleta o ausencia de parche en esta etiqueta.
- La primera ejecución falló por una asunción incorrecta sobre cache vacío; está documentada y no se presenta como evidencia funcional. La segunda es la ejecución válida.
- No se modificó código de producción ni se probó una corrección.

## Siguiente paso automático recomendado

Identificar el commit/release upstream que Babylon atribuye a este hallazgo y ejecutar sin cambios el mismo test contra ese commit y una versión posterior. Comparar estado derivado, cache y tabla completa de poder. Hasta entonces: F-02 reproducida en v4.5.0, duplicada respecto al hallazgo público, regresión histórica no demostrada.