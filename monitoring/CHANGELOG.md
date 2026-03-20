# Changelog

## [0.5.1](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.5.0...monitoring-v0.5.1) (2026-03-20)


### Bug Fixes

* **docker:** use numeric UID for chown in dashboard Dockerfile ([3cabe6a](https://github.com/Audumla/AUDiotMonitor/commit/3cabe6a3eb5abd7ccb62554829126f49a6f61d6f))

## [0.5.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.4.1...monitoring-v0.5.0) (2026-03-20)


### Features

* **monitoring:** add reproducible management scripts ([0c7d6db](https://github.com/Audumla/AUDiotMonitor/commit/0c7d6db260d01ff21aacb8ce7fda2fa1f2d31af0))
* **monitoring:** unify deployment, harden test coverage and improve dashboard operability ([73e0425](https://github.com/Audumla/AUDiotMonitor/commit/73e0425f58cd8ab5a09db25964c277934b6996a4))

## [0.4.1](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.4.0...monitoring-v0.4.1) (2026-03-19)


### Bug Fixes

* **monitoring:** build dashboard image without chmod step ([489ed8a](https://github.com/Audumla/AUDiotMonitor/commit/489ed8a0f223cb4d7df9bb579d54f94b8114dd1a))

## [0.4.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.3.0...monitoring-v0.4.0) (2026-03-19)


### Features

* **monitoring:** add host-owned Prometheus recording rules ([baae978](https://github.com/Audumla/AUDiotMonitor/commit/baae978227e18e275fbc8d7a24f7c43c0a947cdf))

## [0.3.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.2.0...monitoring-v0.3.0) (2026-03-19)


### Features

* **monitoring:** make kiosk dashboard selection config-driven ([50b8dfb](https://github.com/Audumla/AUDiotMonitor/commit/50b8dfbadb2543225ef7d26f4c6e4d1ad6d29d94))


### Bug Fixes

* **monitoring:** correct GPU VRAM dashboard queries ([8b90e52](https://github.com/Audumla/AUDiotMonitor/commit/8b90e52caddafd7513c3746d25f4d3e7b9c8cf19))
* **monitoring:** keep kiosk running without example dashboards ([0fc0bc2](https://github.com/Audumla/AUDiotMonitor/commit/0fc0bc21cf187ced6f9f96230d64f438b283e1f8))
* **monitoring:** remove empty dashboard panels ([b5c88e4](https://github.com/Audumla/AUDiotMonitor/commit/b5c88e4cacb36b7e6236e4b99d42ae002f0c9333))
* **monitoring:** use 1920x440 triple GPU dashboard layout ([60fde0c](https://github.com/Audumla/AUDiotMonitor/commit/60fde0c5df1e0df55f28b3947eafbbcce2986d41))

## [0.2.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.1.0...monitoring-v0.2.0) (2026-03-19)


### Features

* add panel-1920x440-dials dashboard for 3 GPUs + CPU/RAM ([60b0c96](https://github.com/Audumla/AUDiotMonitor/commit/60b0c964ae77a5b467f438ccba75e8cb42de076b))
* **dashboard:** add auto-resolution kiosk launcher and installer ([a8825b9](https://github.com/Audumla/AUDiotMonitor/commit/a8825b9347ea204e953cc5506612575fd57c86e6))
* **dashboard:** add config/ bind-mount layout, simplify RPi compose ([2e30a02](https://github.com/Audumla/AUDiotMonitor/commit/2e30a02dae909890cf63e1b8e1fd005759ed5851))
* **dashboard:** add Infinity datasource and RAPL/systemd panels ([68ef780](https://github.com/Audumla/AUDiotMonitor/commit/68ef780daafd284f46b904aa0f5960ba3d77e481))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))
* **monitoring:** enable all non-default node-exporter collectors ([cd9e2c9](https://github.com/Audumla/AUDiotMonitor/commit/cd9e2c969afca5de4b865a48faeb970a29ac80b7))
* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))
* simplify collector — embed prometheus config in entrypoint, remove init container ([ba03659](https://github.com/Audumla/AUDiotMonitor/commit/ba03659763754cb2f65f1ed789c5b2b268a214d6))


### Bug Fixes

* added default dashboard profiles ([40092d6](https://github.com/Audumla/AUDiotMonitor/commit/40092d677bc57a7826a3b93dbff3fdeb44156f66))
* added more monitoring ([c28ef2c](https://github.com/Audumla/AUDiotMonitor/commit/c28ef2c251565dce7df6aacb1de912ca9d3c3b37))
* calculate explicit VRAM capacity usage percent for GPUs and update dashboard queries ([61d07ec](https://github.com/Audumla/AUDiotMonitor/commit/61d07ec74b201332018f0b2463441941db71d5ab))
* embed Prometheus config inline — remove init container ([d00ec3b](https://github.com/Audumla/AUDiotMonitor/commit/d00ec3b30dca17dd0333fae6d690fef3fd21d8fd))
* gpu metics ([72e5153](https://github.com/Audumla/AUDiotMonitor/commit/72e515361898fc09233c44742de6016366f2db53))
* make grafana-init idempotent — skip downloads if files already exist ([dbd0dc0](https://github.com/Audumla/AUDiotMonitor/commit/dbd0dc0ecdf81f2acdf61d3040cbd19d74c16b1f))
* **monitoring:** node-exporter host networking, dbus, udev mounts ([60ebdc6](https://github.com/Audumla/AUDiotMonitor/commit/60ebdc6830b896a09227bb7a12c127ca63551a09))
* remove build context from collector compose — use published Docker Hub image ([4861d88](https://github.com/Audumla/AUDiotMonitor/commit/4861d8849cea62eeebcbfc513cc5b35c1d50e890))
* update collector compose with working SELinux and permission config ([45ba571](https://github.com/Audumla/AUDiotMonitor/commit/45ba5714bbe14cb0e5f7cbc193093ca71d8ece91))
* update kiosk startup script ([73fe9e5](https://github.com/Audumla/AUDiotMonitor/commit/73fe9e57cb1654fe94bc3beeb3f9a67171143f40))
* update monitoring capabilities ([0e9e610](https://github.com/Audumla/AUDiotMonitor/commit/0e9e61006149ec576e609aa4b2c5b8357d54d192))
* updated extension monitoring ([444f052](https://github.com/Audumla/AUDiotMonitor/commit/444f05267d93cd17688c1799e28a8f80f272d7ea))
