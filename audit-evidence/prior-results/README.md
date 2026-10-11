# Evidencia preservada de ejecuciones anteriores

Esta carpeta conserva copias byte por byte de los informes y artefactos de las ejecuciones previas; no se reemplazaron los archivos originales.

## Resultados ya verificados en ejecuciones anteriores

- v4.5.0 (`f9006e4232dece8050b08556c4617bc0aaa07128`): el control pasó; el test objetivo alcanzó su aserción final y observó estado derivado `EXPIRED` con `2.000.000.000` sats de poder persistido.
- upstream main (`c2f731f4531b2b22afe5047f152e108b944bf811`): mismo resultado observado.
- Enlaces de ejecución: [run 38098321795](https://github.com/arkakill/babylon/actions/runs/38098321795), [run 38098547912](https://github.com/arkakill/babylon/actions/runs/38098547912).

Estos resultados anteriores no se cuentan como ejecuciones nuevas de esta matriz.

## Intento histórico de la matriz

[Run 38100558671](https://github.com/arkakill/babylon/actions/runs/38100558671) encontró incompatibilidad entre los helpers de los harnesses históricos antes de llegar a las aserciones finales para los casos observados. Por ese motivo no se interpreta como prueba de que F-02 se reproduzca o no en esos SHAs. El workflow adaptado está en la rama de auditoría y se ejecutará por separado en el repositorio privado.
