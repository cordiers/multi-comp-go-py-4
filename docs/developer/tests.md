# Tests

The principles about tests on n0stack.

## Test size

- according to https://testing.googleblog.com/2010/12/test-sizes.html

### small

- unit test about logic
- integration test about side effect
- without side effect, for example...
    - persistent data
    - control middleware
- 副作用は agent に固まっているので、 agent だけモックすることで...
    - ロジックの結合テストを small にて行える
    - agent からロジックを消すことで分散耐性を向上
    - モックの開発工数を減らせる

#### Goal

- coverage n0core/pkg/api without agent > 70 %
- coverage n0core/pkg/api with agent > 50 %
- coverage n0core/pkg/datastore/memory > 70 %

### medium

- integration test about side effect on standalone
- gRPC fuzzing about logic

#### Goal

- coverage n0core/pkg/api > 70 %
- coverage n0core/pkg/datastore/etcd > 80 %
- coverage n0core/pkg/driver > 60 %

### large

- E2E

## TODO

- [x] 現状のテストが通るようにする
- [x] 各 API のモックの作成と差し替え
- [x] Agent からロジックの切り出し
- [x] Agent のモック作成
- [x] medium -> small に
- [ ] API のテストを書いていく
