# Changelog

## [0.21.2](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.21.1...hwexp-v0.21.2) (2026-03-28)


### Bug Fixes

* **mapping:** restore GPU VRAM series and enforce exact raw_name matching ([4bfbb35](https://github.com/Audumla/AUDiotMonitor/commit/4bfbb3524e25d8d7f0434ae3aa29afc7b8298e98))

## [0.21.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.21.0...hwexp-v0.21.1) (2026-03-28)


### Bug Fixes

* **release:** align adapters, compose mounts, and docs for artifact-only deploy ([2c86408](https://github.com/Audumla/AUDiotMonitor/commit/2c864080fdcd8b12f8b54024b160483f3df521f8))

## [0.21.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.20.0...hwexp-v0.21.0) (2026-03-28)


### Features

* **hwexp:** add manifest, storage, and system adapters with correlation ([1dd089a](https://github.com/Audumla/AUDiotMonitor/commit/1dd089a6f702181d0e8f380c6fc48f332085927a))


### Bug Fixes

* minor updates ([e3e3098](https://github.com/Audumla/AUDiotMonitor/commit/e3e3098dcccf3b4c0a8ce4ca8aabdfa4dd8ca40d))

## [0.20.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.19.2...hwexp-v0.20.0) (2026-03-19)


### Features

* **monitoring:** add host-owned Prometheus recording rules ([baae978](https://github.com/Audumla/AUDiotMonitor/commit/baae978227e18e275fbc8d7a24f7c43c0a947cdf))

## [0.19.2](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.19.1...hwexp-v0.19.2) (2026-03-19)


### Bug Fixes

* calculate explicit VRAM capacity usage percent for GPUs and update dashboard queries ([61d07ec](https://github.com/Audumla/AUDiotMonitor/commit/61d07ec74b201332018f0b2463441941db71d5ab))

## [0.19.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.19.0...hwexp-v0.19.1) (2026-03-19)


### Bug Fixes

* gpu metics ([72e5153](https://github.com/Audumla/AUDiotMonitor/commit/72e515361898fc09233c44742de6016366f2db53))

## [0.19.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.18.0...hwexp-v0.19.0) (2026-03-19)


### Features

* **httpapi:** add root index page listing all available endpoints ([ebb2adb](https://github.com/Audumla/AUDiotMonitor/commit/ebb2adbd4225e7552573ed37e206fc96df68bb7e))

## [0.18.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.17.2...hwexp-v0.18.0) (2026-03-19)


### Features

* **hwexp:** add /debug/state consolidated JSON endpoint ([e91c095](https://github.com/Audumla/AUDiotMonitor/commit/e91c0951683ee9d1005274fcfd0039f67a5d4dc6))


### Bug Fixes

* **hwexp:** show linux_static in startup banner adapter list ([888e59b](https://github.com/Audumla/AUDiotMonitor/commit/888e59b3f77e19d97c4243b6742f837f4beaa66b))

## [0.17.2](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.17.1...hwexp-v0.17.2) (2026-03-19)


### Bug Fixes

* **hwexp:** classify iwlwifi/ath/mt76/rtw hwmon devices as network/wifi ([dbe1826](https://github.com/Audumla/AUDiotMonitor/commit/dbe18268fccde9de8c25efdf922b76dc7847c4cb))

## [0.17.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.17.0...hwexp-v0.17.1) (2026-03-19)


### Bug Fixes

* **hwexp:** WiFi hwmon classification, SPD slot labeling, NIC vendor enrichment ([b86befe](https://github.com/Audumla/AUDiotMonitor/commit/b86befe8d7dbf3486b0e25f0cc7cad08d9ec56c6))

## [0.17.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.16.0...hwexp-v0.17.0) (2026-03-19)


### Features

* **hwexp:** add pcidb enrichment, CPU cache, DIMM discovery, GPU metrics ([1ee9d7d](https://github.com/Audumla/AUDiotMonitor/commit/1ee9d7d73222d4b631b077e038f1a8fe01306155))


### Bug Fixes

* **hwexp:** wireless device classification, DMI placeholder filtering, NVMe vendors ([62f4493](https://github.com/Audumla/AUDiotMonitor/commit/62f4493201be5c23d07a64103aa4cf09aab6c406))

## [0.16.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.15.0...hwexp-v0.16.0) (2026-03-19)


### Features

* **hwexp:** expand hardware discovery and metrics collection ([9bb74f6](https://github.com/Audumla/AUDiotMonitor/commit/9bb74f6b11473cc20345393bd8f601ea38396eb8))

## [0.15.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.14.1...hwexp-v0.15.0) (2026-03-19)


### Features

* **hwexp:** improve device classification and model enrichment ([e9a28c8](https://github.com/Audumla/AUDiotMonitor/commit/e9a28c88523719eb0df950b0ac70d7c76617bea8))


### Bug Fixes

* **hwexp:** normalise startup banner line widths ([152e2ca](https://github.com/Audumla/AUDiotMonitor/commit/152e2cac21c02b3921f9a6aa661cfaf51e1614c8))
* mappings ([2b819eb](https://github.com/Audumla/AUDiotMonitor/commit/2b819ebb967718c8a074499cd8d94e82a6c01b97))

## [0.14.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.14.0...hwexp-v0.14.1) (2026-03-19)


### Performance Improvements

* **ci:** cross-compile in Docker builder, parallelize docker-publish job ([af84fd8](https://github.com/Audumla/AUDiotMonitor/commit/af84fd888c8d1ad006d587ee65452539db29afef))

## [0.14.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.13.2...hwexp-v0.14.0) (2026-03-19)


### Features

* **hwexp:** print startup banner with version, config, and adapter summary ([889f701](https://github.com/Audumla/AUDiotMonitor/commit/889f701a2c8444b5c0412201735367d3d6132c76))


### Bug Fixes

* added version output on start ([2c32ad1](https://github.com/Audumla/AUDiotMonitor/commit/2c32ad1b465421b5384c340858cba436c9179901))

## [0.13.2](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.13.1...hwexp-v0.13.2) (2026-03-19)


### Bug Fixes

* added more monitoring ([c28ef2c](https://github.com/Audumla/AUDiotMonitor/commit/c28ef2c251565dce7df6aacb1de912ca9d3c3b37))
* update monitoring capabilities ([0e9e610](https://github.com/Audumla/AUDiotMonitor/commit/0e9e61006149ec576e609aa4b2c5b8357d54d192))
* updated extension monitoring ([444f052](https://github.com/Audumla/AUDiotMonitor/commit/444f05267d93cd17688c1799e28a8f80f272d7ea))

## [0.13.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.13.0...hwexp-v0.13.1) (2026-03-19)


### Bug Fixes

* remove configs/ from .dockerignore so Dockerfile can COPY default config ([53ab113](https://github.com/Audumla/AUDiotMonitor/commit/53ab113ffcbd985a4b66db1d881ac789ad0663c0))

## [0.13.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.12.0...hwexp-v0.13.0) (2026-03-19)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))
* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))
* **tests:** update all integration tests to handle hwexp_up host label ([b9119f3](https://github.com/Audumla/AUDiotMonitor/commit/b9119f36a2e12cc378f0a3593ea41323fee9c4d5))

## [0.12.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.11.0...hwexp-v0.12.0) (2026-03-19)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))
* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))
* **tests:** update all integration tests to handle hwexp_up host label ([b9119f3](https://github.com/Audumla/AUDiotMonitor/commit/b9119f36a2e12cc378f0a3593ea41323fee9c4d5))

## [0.11.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.10.0...hwexp-v0.11.0) (2026-03-19)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))
* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))
* **tests:** update all integration tests to handle hwexp_up host label ([b9119f3](https://github.com/Audumla/AUDiotMonitor/commit/b9119f36a2e12cc378f0a3593ea41323fee9c4d5))

## [0.10.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.9.0...hwexp-v0.10.0) (2026-03-19)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))
* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))
* **tests:** update all integration tests to handle hwexp_up host label ([b9119f3](https://github.com/Audumla/AUDiotMonitor/commit/b9119f36a2e12cc378f0a3593ea41323fee9c4d5))

## [0.9.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.8.0...hwexp-v0.9.0) (2026-03-19)


### Features

* self-contained deploy — no repo clone needed ([69b899c](https://github.com/Audumla/AUDiotMonitor/commit/69b899cd42c5dedc1fb9e4ff93ae69117a08b85a))

## [0.8.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.7.1...hwexp-v0.8.0) (2026-03-19)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))
* **tests:** update all integration tests to handle hwexp_up host label ([b9119f3](https://github.com/Audumla/AUDiotMonitor/commit/b9119f36a2e12cc378f0a3593ea41323fee9c4d5))

## [0.7.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.7.0...hwexp-v0.7.1) (2026-03-19)


### Bug Fixes

* **tests:** update all integration tests to handle hwexp_up host label ([b9119f3](https://github.com/Audumla/AUDiotMonitor/commit/b9119f36a2e12cc378f0a3593ea41323fee9c4d5))

## [0.7.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.6.0...hwexp-v0.7.0) (2026-03-18)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))
* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))
* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))

## [0.6.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.5.1...hwexp-v0.6.0) (2026-03-18)


### Features

* **hwexp:** restructure deployment into collector/dashboard stacks ([ac81632](https://github.com/Audumla/AUDiotMonitor/commit/ac816323c6e5ff5370d6adda03f1518fb4ff945a))

## [0.5.1](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.5.0...hwexp-v0.5.1) (2026-03-18)


### Bug Fixes

* documentation updates ([e5df5ff](https://github.com/Audumla/AUDiotMonitor/commit/e5df5ffb9c7d2785f8275cb58126edc7d24eeee5))

## [0.5.0](https://github.com/Audumla/AUDiotMonitor/compare/hwexp-v0.4.0...hwexp-v0.5.0) (2026-03-18)


### Features

* **hwexp:** auto-map hardware sensors on first run, grafana dashboards and install guide ([d7b4da0](https://github.com/Audumla/AUDiotMonitor/commit/d7b4da0d7a296edc03bcc260dd8a35313dd8ad90))

## [0.4.0](https://github.com/Audumla/AUDiotMonitor/compare/v0.3.0...v0.4.0) (2026-03-18)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))


### Bug Fixes

* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))

## [0.3.0](https://github.com/Audumla/AUDiotMonitor/compare/v0.2.0...v0.3.0) (2026-03-18)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* test failure on docker deploy ([ce77a60](https://github.com/Audumla/AUDiotMonitor/commit/ce77a60dbf7f69c129cf22d685b8a5dec6ae249a))

## [0.2.0](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.10...v0.2.0) (2026-03-18)


### Features

* containerise full stack — hwexp + node-exporter + prometheus + grafana ([2e50d35](https://github.com/Audumla/AUDiotMonitor/commit/2e50d358fc1bef300d99e455ce73dc17bba37b7d))
* docker enabling ([5ec0286](https://github.com/Audumla/AUDiotMonitor/commit/5ec0286b9256f9f21e4f7b86ad27b70bf45adb8e))


### Bug Fixes

* ensure first-time install works correctly on fresh Linux hosts ([f2f1732](https://github.com/Audumla/AUDiotMonitor/commit/f2f173288e558beef95f55891aa0b62a0ee599ee))
* open/close firewall port 9200 on install and remove ([14391d1](https://github.com/Audumla/AUDiotMonitor/commit/14391d1f22b0047b1f84be60463a8fed65ce0a6b))

## [0.1.10](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.9...v0.1.10) (2026-03-18)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.9](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.8...v0.1.9) (2026-03-18)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.8](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.7...v0.1.8) (2026-03-18)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* hardware monitoring ([0e3f46f](https://github.com/Audumla/AUDiotMonitor/commit/0e3f46f4bd4a84d992f590c4228aaa0655292a92))
* host issue ([8ea5efc](https://github.com/Audumla/AUDiotMonitor/commit/8ea5efce371c9d9235fc1af77de10ee8e42e10b8))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.7](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.6...v0.1.7) (2026-03-17)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* release builds ([c719b21](https://github.com/Audumla/AUDiotMonitor/commit/c719b2112db8c43914c37f601708c87c7f267a9f))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.6](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.5...v0.1.6) (2026-03-17)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* cross platform testing ([6d3b0e2](https://github.com/Audumla/AUDiotMonitor/commit/6d3b0e20def70f67db539ed33ee19d9ee9b22eec))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.5](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.4...v0.1.5) (2026-03-17)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.4](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.3...v0.1.4) (2026-03-17)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* release fixes ([fe0321d](https://github.com/Audumla/AUDiotMonitor/commit/fe0321d8e53d8f340ad2126ca3c059d4b29f168d))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.3](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.2...v0.1.3) (2026-03-17)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.2](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.1...v0.1.2) (2026-03-17)


### Bug Fixes

* agent release errors ([46b21d1](https://github.com/Audumla/AUDiotMonitor/commit/46b21d118bbf883c8c62765ea38cba62bca41715))
* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))

## [0.1.1](https://github.com/Audumla/AUDiotMonitor/compare/v0.1.0...v0.1.1) (2026-03-17)


### Bug Fixes

* release issues ([a370303](https://github.com/Audumla/AUDiotMonitor/commit/a370303ac37dcd2369612f1301e675e9b205c490))
