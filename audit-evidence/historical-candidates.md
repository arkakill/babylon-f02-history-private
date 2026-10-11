# Reconstrucción histórica de PR #631 y PR #634

## Evidencia directa

- [PR #631](https://github.com/babylonlabs-io/babylon/pull/631): padre `cbc3a9a8c0c1e143f44f819dc145051dabf0f122`; commit de corrección `6d9b66816d33bd5bbb2abc9a55eb29324d836715`.
- [PR #634](https://github.com/babylonlabs-io/babylon/pull/634): backport a `release/v1.x`; padre `b68ef264564cfb01dd0d270076cec4055f3d1c6d`; commit `1ebc3727fba1bf6231cb11d70554b12fa58d2342`.
- El cambio de #631 condiciona el procesamiento de la rama de expiración de poder de voto a que la delegación tenga quórum covenant. La intención documentada es evitar restar poder de una delegación expirada que nunca se activó y cuyo poder nunca fue añadido.
- #634 replica el cambio en la rama de backport.

## Comparación con F-02 y OpenZeppelin

El hallazgo de OpenZeppelin de 2026 describe una secuencia de frontera de estado en la que una delegación aún pendiente por falta de quórum puede expirar y después quedar activa cuando llegan firmas tardías. F-02 prueba el keeper de integración, la cola de eventos, el estado derivado y la distribución almacenada tras consumir `EXPIRED`.

La coincidencia técnica es sustancial, pero las PR #631 y #634 no citan explícitamente el informe de OpenZeppelin ni permiten atribuirles esa remediación. La secuencia multi-bloque debe ejecutarse contra el padre y la corrección antes de clasificar una corrección como verificada o incompleta. Los resultados de esa comparación se incorporarán cuando la matriz de Actions termine.

Fuente adicional: [OpenZeppelin, State Changes at the Boundary](https://www.openzeppelin.com/news/state-changes-at-the-boundary-lessons-from-security-research-on-babylon).

## Límite actual

No se atribuye causalidad ni se declara regresión por fechas. Si un harness no compila o el test no llega a la aserción final, ese SHA será `INCONCLUSA` para la comparación histórica.


## Verificación adicional de commits y atribución

La API oficial de GitHub confirmó las relaciones de padres examinadas en la matriz: `6d9b66816d33bd5bbb2abc9a55eb29324d836715` tiene como único padre `cbc3a9a8c0c1e143f44f819dc145051dabf0f122`; `1ebc3727fba1bf6231cb11d70554b12fa58d2342` tiene como único padre `b68ef264564cfb01dd0d270076cec4055f3d1c6d`. El commit de PR #631 se tituló “ignore expire events when there is no covenant quorum”; su descripción dice que una delegación sin quórum nunca se activó y sus sats no se contaron al proveedor. El cuerpo de PR #634 identifica expresamente su cambio como backport de PR #631 a `release/v1.x`.

El cambio de PR #631 evita ejecutar la rama de expiración que procesa el poder cuando falta el quórum covenant. Su prueba añadida procesa una expiración sin quórum; no cubre la secuencia de completar el quórum después de que el evento ya se consumió. Los documentos públicos consultados no contienen una atribución explícita a OpenZeppelin. Por ello, incluso si la matriz demuestra que el mismo escenario sigue fallando después del cambio, la descripción precisa será “corrección incompleta para esta secuencia” y no “regresión” ni una atribución de origen no documentada.

Referencias verificables: [PR #631](https://github.com/babylonlabs-io/babylon/pull/631), [PR #634](https://github.com/babylonlabs-io/babylon/pull/634), [commit #631](https://github.com/babylonlabs-io/babylon/commit/6d9b66816d33bd5bbb2abc9a55eb29324d836715), [commit #634](https://github.com/babylonlabs-io/babylon/commit/1ebc3727fba1bf6231cb11d70554b12fa58d2342).
