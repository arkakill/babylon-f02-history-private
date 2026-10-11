# Lectura del primer intento de matriz histórica

- Run: https://github.com/arkakill/babylon/actions/runs/38100558671
- Commit del workflow: `02a6083bc7a7a4b2223ee93914edea56bc04b98f`.
- El resumen de GitHub muestra seis trabajos completados y estado global Success, pero cuatro revisiones históricas reportan explícitamente `Control: FAIL_OR_BLOCKED`, `Target: FAILED_BEFORE_FINAL_INVARIANT_OR_COMPILE` y `Final assertion: NOT_VERIFIED`. Por tanto ese estado global no constituye evidencia sobre F-02 en esos cuatro SHA.
- Para `v4.5.0` (`f9006e4232dece8050b08556c4617bc0aaa07128`), el resumen sí registra control PASS (código 0) y F-02 reproducida en la aserción final (código 1). Los reportes previos archivados en `prior-results/` contienen el estado y el poder observados.

## Errores de compatibilidad anotados por GitHub

La salida de compilación del primer intento documentó estas diferencias del harness antiguo:

- `NewHelper` aceptaba menos argumentos y el helper no exponía `FinalityHooks`.
- `GenerateCovenantSignaturesMessages` necesitaba el mensaje de creación de la delegación.
- El keeper antiguo no tenía `BtcDelStatus`; la delegación ofrecía `GetStatus(btcHeight, covenantQuorum)`.
- En las revisiones de PR #631, `CreateDelegationWithBtcBlockHeight` esperaba un slice de claves de finality providers.
- Al retirar el mock de hooks, el import `ftypes` quedaba sin uso.

La adaptación de la matriz modificada está acotada a esos nombres, argumentos, import y tipo de proveedor. Conserva la secuencia de bloques, el procesamiento real del keeper, las aserciones de eventos, el estado, la caché y la tabla persistida. No cambia código de producción ni dependencias. La verificación con la nueva matriz aún está pendiente de ejecución en Actions.

## Fuente del detalle

La página del run y sus anotaciones están enlazadas arriba. En el run, v4.5.0 muestra Go `1.23.8 linux/amd64`; los resúmenes históricos visibles muestran Go `1.23.1 linux/amd64`.
