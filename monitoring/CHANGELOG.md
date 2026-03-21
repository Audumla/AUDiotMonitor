# Changelog

## [0.8.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.7.3...monitoring-v0.8.0) (2026-03-21)


### Features

* **kiosk:** add KIOSK_IDLE_TIMEOUT for DPMS wake-on-touch support ([5f98444](https://github.com/Audumla/AUDiotMonitor/commit/5f984441b4b6b02cb1305805c4960179a276c834))
* **kiosk:** expose advanced DPMS states to allow screen dimming ([c4b4300](https://github.com/Audumla/AUDiotMonitor/commit/c4b4300bc74d01cfeba04959a703889b83f56842))
* **logs:** add timestamps to audiot-dashboard and kiosk logs ([85109c0](https://github.com/Audumla/AUDiotMonitor/commit/85109c0d6d82428726c1918560d52c4d77589c93))


### Bug Fixes

* **dashboard:** resolve self-update loop, wrong install dir, and kiosk permission errors ([6bbd69a](https://github.com/Audumla/AUDiotMonitor/commit/6bbd69af06932dc1ebc505751acd576cdfe6f919))
* **kiosk:** improve wait_for_grafana timeout and logging ([baf4dce](https://github.com/Audumla/AUDiotMonitor/commit/baf4dce39a659ff0b586130ea7031b737f614982))
* **kiosk:** remove wait_for_grafana to speed up browser startup and cleanup script ([a96dfb2](https://github.com/Audumla/AUDiotMonitor/commit/a96dfb2497cefe2f583d9ce9e71ee779c0435f07))
* **kiosk:** use /public/img/ for custom CSS to ensure Grafana serves it ([d8a7056](https://github.com/Audumla/AUDiotMonitor/commit/d8a7056271d6bda9322611c24805e52b31a8fa40))
* **kiosk:** use absolute path for CSS and update kiosk URL parameters ([30f6a86](https://github.com/Audumla/AUDiotMonitor/commit/30f6a860dc0774a9f4089a329104e9853db2957e))
* **script:** add cache-buster to self-update URL ([7a4a941](https://github.com/Audumla/AUDiotMonitor/commit/7a4a941a511f25bcb7c679952c41164cb4877fd0))
* **script:** auto-detect INSTALL_DIR if run from a dashboard folder ([e98ffeb](https://github.com/Audumla/AUDiotMonitor/commit/e98ffeb01b2a2cb4ee35d7cb42082de4285c558b))
* **script:** ensure TARGET_USER detection in do_restart_kiosk to fix log permissions ([6e74d74](https://github.com/Audumla/AUDiotMonitor/commit/6e74d74dd625bac1065dd35c77005a76e44ec1ae))
* **script:** make self-update line-ending agnostic and resume operation after update ([972c778](https://github.com/Audumla/AUDiotMonitor/commit/972c778a77b3912e38e43609d610847a82ac5502))
* **script:** make UI patch script more robust and fail-safe ([bbcaccf](https://github.com/Audumla/AUDiotMonitor/commit/bbcaccf105d4e4722c5f5516d4ce32747ac3d845))
* **script:** push uncommitted changes for UI patching ([de776a7](https://github.com/Audumla/AUDiotMonitor/commit/de776a7e42f22a89c5a6dfe0c2c35d6d1c4f4793))
* **script:** resolve log file permission issues during kiosk restart ([5b16a58](https://github.com/Audumla/AUDiotMonitor/commit/5b16a5848e3f9bb463289ad08c1de87db426fdb9))
* **script:** use checksums for self-update to prevent loop ([df6670a](https://github.com/Audumla/AUDiotMonitor/commit/df6670aeaff340fb9c27bc0c639b3516cd273a93))
* **script:** use more robust sed pattern for CSS injection ([cdf47c4](https://github.com/Audumla/AUDiotMonitor/commit/cdf47c4d640d2e9c8d5edf938a687b602ec06b21))

## [0.7.3](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.7.2...monitoring-v0.7.3) (2026-03-20)


### Bug Fixes

* remove footer ([97da90e](https://github.com/Audumla/AUDiotMonitor/commit/97da90e9e759e61da627f5badf03897536ac302c))
* **script:** correct non-portable mkdir command in installer ([8155bf9](https://github.com/Audumla/AUDiotMonitor/commit/8155bf9733bc6cbf37f26fe2e434c4d52e39dcec))
* **script:** harden sudo environment passing with env command ([8e1e5cf](https://github.com/Audumla/AUDiotMonitor/commit/8e1e5cfa763746d7aee381bc76333077610283e9))
* **script:** improve sudo handling for systemd user service installation ([15937a7](https://github.com/Audumla/AUDiotMonitor/commit/15937a753b6efde5eb86fec46003072e91711f28))

## [0.7.2](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.7.1...monitoring-v0.7.2) (2026-03-20)


### Bug Fixes

* **kiosk:** implement user-provided URL for true header-less mode ([e8d061f](https://github.com/Audumla/AUDiotMonitor/commit/e8d061fc142177caeb42e591c9f8d79b769f957a))
* **kiosk:** make kiosk restart logic more robust ([81da0f9](https://github.com/Audumla/AUDiotMonitor/commit/81da0f9fba179867e65bce60a3b99cd7c0e92920))
* **kiosk:** simplify kiosk-install.sh to only manage autostart ([86630d3](https://github.com/Audumla/AUDiotMonitor/commit/86630d3a62e20c9bfdd67f3319b54eab578957d2))
* **kiosk:** use kiosk=tv parameter for true header-less mode per official docs ([2905fd6](https://github.com/Audumla/AUDiotMonitor/commit/2905fd6636392022f7c2e884298f11ab20b17f00))

## [0.7.1](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.7.0...monitoring-v0.7.1) (2026-03-20)


### Bug Fixes

* **collector:** restore docker GPU metric config ([0808110](https://github.com/Audumla/AUDiotMonitor/commit/08081109a9b8d70932c8e09257ab0cd63cca5164))
* enforce relative config paths ([ba18b11](https://github.com/Audumla/AUDiotMonitor/commit/ba18b1137930198cbc3a6c504ea987b92dfb6925))

## [0.7.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.6.0...monitoring-v0.7.0) (2026-03-20)


### Features

* add panel-1920x440-dials dashboard for 3 GPUs + CPU/RAM ([60b0c96](https://github.com/Audumla/AUDiotMonitor/commit/60b0c964ae77a5b467f438ccba75e8cb42de076b))
* **dashboard:** add auto-resolution kiosk launcher and installer ([a8825b9](https://github.com/Audumla/AUDiotMonitor/commit/a8825b9347ea204e953cc5506612575fd57c86e6))
* **dashboard:** add config/ bind-mount layout, simplify RPi compose ([2e30a02](https://github.com/Audumla/AUDiotMonitor/commit/2e30a02dae909890cf63e1b8e1fd005759ed5851))
* **dashboard:** add Infinity datasource and RAPL/systemd panels ([68ef780](https://github.com/Audumla/AUDiotMonitor/commit/68ef780daafd284f46b904aa0f5960ba3d77e481))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))
* **metrics:** generate side-by-side GPU metrics for compatibility and future use ([82e5881](https://github.com/Audumla/AUDiotMonitor/commit/82e5881c7ef4667d0e253b19f9b834b672bc469e))
* **monitoring:** add host-owned Prometheus recording rules ([baae978](https://github.com/Audumla/AUDiotMonitor/commit/baae978227e18e275fbc8d7a24f7c43c0a947cdf))
* **monitoring:** add reproducible management scripts ([0c7d6db](https://github.com/Audumla/AUDiotMonitor/commit/0c7d6db260d01ff21aacb8ce7fda2fa1f2d31af0))
* **monitoring:** enable all non-default node-exporter collectors ([cd9e2c9](https://github.com/Audumla/AUDiotMonitor/commit/cd9e2c969afca5de4b865a48faeb970a29ac80b7))
* **monitoring:** make kiosk dashboard selection config-driven ([50b8dfb](https://github.com/Audumla/AUDiotMonitor/commit/50b8dfbadb2543225ef7d26f4c6e4d1ad6d29d94))
* **monitoring:** unify deployment, harden test coverage and improve dashboard operability ([73e0425](https://github.com/Audumla/AUDiotMonitor/commit/73e0425f58cd8ab5a09db25964c277934b6996a4))
* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))
* simplify collector — embed prometheus config in entrypoint, remove init container ([ba03659](https://github.com/Audumla/AUDiotMonitor/commit/ba03659763754cb2f65f1ed789c5b2b268a214d6))


### Bug Fixes

* added default dashboard profiles ([40092d6](https://github.com/Audumla/AUDiotMonitor/commit/40092d677bc57a7826a3b93dbff3fdeb44156f66))
* added more monitoring ([c28ef2c](https://github.com/Audumla/AUDiotMonitor/commit/c28ef2c251565dce7df6aacb1de912ca9d3c3b37))
* calculate explicit VRAM capacity usage percent for GPUs and update dashboard queries ([61d07ec](https://github.com/Audumla/AUDiotMonitor/commit/61d07ec74b201332018f0b2463441941db71d5ab))
* **docker:** use numeric UID for chown in dashboard Dockerfile ([3cabe6a](https://github.com/Audumla/AUDiotMonitor/commit/3cabe6a3eb5abd7ccb62554829126f49a6f61d6f))
* embed Prometheus config inline — remove init container ([d00ec3b](https://github.com/Audumla/AUDiotMonitor/commit/d00ec3b30dca17dd0333fae6d690fef3fd21d8fd))
* gpu metics ([72e5153](https://github.com/Audumla/AUDiotMonitor/commit/72e515361898fc09233c44742de6016366f2db53))
* make grafana-init idempotent — skip downloads if files already exist ([dbd0dc0](https://github.com/Audumla/AUDiotMonitor/commit/dbd0dc0ecdf81f2acdf61d3040cbd19d74c16b1f))
* **metrics:** restore original GPU metric names for dashboard compatibility ([d4ffcf6](https://github.com/Audumla/AUDiotMonitor/commit/d4ffcf6f71a9c3fd407f2894ed6387309ae81abe))
* **metrics:** revert custom rule generation to not interfere with raw metrics ([543905c](https://github.com/Audumla/AUDiotMonitor/commit/543905c64981e87fb76970e97c73bc5af2cb58b8))
* **monitoring:** build dashboard image without chmod step ([489ed8a](https://github.com/Audumla/AUDiotMonitor/commit/489ed8a0f223cb4d7df9bb579d54f94b8114dd1a))
* **monitoring:** correct GPU VRAM dashboard queries ([8b90e52](https://github.com/Audumla/AUDiotMonitor/commit/8b90e52caddafd7513c3746d25f4d3e7b9c8cf19))
* **monitoring:** keep kiosk running without example dashboards ([0fc0bc2](https://github.com/Audumla/AUDiotMonitor/commit/0fc0bc21cf187ced6f9f96230d64f438b283e1f8))
* **monitoring:** node-exporter host networking, dbus, udev mounts ([60ebdc6](https://github.com/Audumla/AUDiotMonitor/commit/60ebdc6830b896a09227bb7a12c127ca63551a09))
* **monitoring:** remove empty dashboard panels ([b5c88e4](https://github.com/Audumla/AUDiotMonitor/commit/b5c88e4cacb36b7e6236e4b99d42ae002f0c9333))
* **monitoring:** use 1920x440 triple GPU dashboard layout ([60fde0c](https://github.com/Audumla/AUDiotMonitor/commit/60fde0c5df1e0df55f28b3947eafbbcce2986d41))
* remove build context from collector compose — use published Docker Hub image ([4861d88](https://github.com/Audumla/AUDiotMonitor/commit/4861d8849cea62eeebcbfc513cc5b35c1d50e890))
* update collector compose with working SELinux and permission config ([45ba571](https://github.com/Audumla/AUDiotMonitor/commit/45ba5714bbe14cb0e5f7cbc193093ca71d8ece91))
* update kiosk startup script ([73fe9e5](https://github.com/Audumla/AUDiotMonitor/commit/73fe9e57cb1654fe94bc3beeb3f9a67171143f40))
* update monitoring capabilities ([0e9e610](https://github.com/Audumla/AUDiotMonitor/commit/0e9e61006149ec576e609aa4b2c5b8357d54d192))
* updated extension monitoring ([444f052](https://github.com/Audumla/AUDiotMonitor/commit/444f05267d93cd17688c1799e28a8f80f272d7ea))

## [0.6.0](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.5.2...monitoring-v0.6.0) (2026-03-20)


### Features

* **metrics:** generate side-by-side GPU metrics for compatibility and future use ([82e5881](https://github.com/Audumla/AUDiotMonitor/commit/82e5881c7ef4667d0e253b19f9b834b672bc469e))

## [0.5.2](https://github.com/Audumla/AUDiotMonitor/compare/monitoring-v0.5.1...monitoring-v0.5.2) (2026-03-20)


### Bug Fixes

* **metrics:** restore original GPU metric names for dashboard compatibility ([d4ffcf6](https://github.com/Audumla/AUDiotMonitor/commit/d4ffcf6f71a9c3fd407f2894ed6387309ae81abe))

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
