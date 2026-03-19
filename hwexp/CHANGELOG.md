# Changelog

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
