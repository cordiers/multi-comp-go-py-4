# Tools

## Circle CI

- バージョンのインクリメント

## Travis CI

- test-small と一応すべてのビルドができるか確認している

## Go Report Card

- Go の静的解析に用いている

## CODE CLIMATE

- とりあえずやっているが、Golangだけの現状では大して効果がなく、適当に残してあるだけ
  - JS などのコンポーネントもできると思うので、そのときに改めて考える
- protobuf の自動生成されたコードは除外設定を行っている

## FOSSA

- ライセンスの確認を行っており、パスするようにする
- etcd と zap (logger) を Ignore の設定をしている
  - 一般的に使われており、深さ 1 では問題ないため暫定処置
