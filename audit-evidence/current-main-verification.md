# Verificación inmutable de upstream main

Fecha de verificación: 2026-10-10.

Comando ejecutado:

```text
git -c http.sslBackend=openssl ls-remote https://github.com/babylonlabs-io/babylon.git refs/heads/main
```

Salida exacta:

```text
c2f731f4531b2b22afe5047f152e108b944bf811 refs/heads/main
```

La matriz usa este SHA completo e inmutable. Es el mismo commit que ya se probó como main histórico; se ejecutará una sola vez para evitar repetir el mismo caso. La referencia móvil `main` no se usa como checkout de la matriz.
