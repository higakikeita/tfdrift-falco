# デモ用 Terraform 構成

`main.tf` を**実際に apply して**デモを回す。フィクスチャの state ではなく、
「書いた → デプロイした → 実機がある」という本物の鎖にするため。

## セットアップ（今夜・1回だけ）

```bash
cd ~/oss-portfolio/talks/oss-summit-korea-2026/tf   # or wherever you keep it
export AWS_PROFILE=draios-dev-developer
export AWS_REGION=ap-northeast-1

terraform init
terraform plan          # 作られるものを確認
terraform apply         # 実機にデプロイ

terraform output security_group_id      # → sg-xxxxxxxx
```

出た SG ID を控える。以降これを使う。

## scan をこの構成に向ける

`demo/scan-config.yaml` の state パスを、この構成の state に変える:

```yaml
providers:
  aws:
    enabled: true
    regions: ["ap-northeast-1"]
    state:
      backend: "local"
      local_path: "REPLACED_BY_SCAN_SH"     # ← scan.sh が絶対パスに書き換える
```

`scan.sh` 側の `sed` の向き先を、フィクスチャ（`scan-state.tfstate`）ではなく
この構成の `terraform.tfstate` にする:

```bash
# demo/scan.sh
sed "s#REPLACED_BY_SCAN_SH#$HERE/../tf/terraform.tfstate#" ...
```

**期待される効果**: `unmanaged` がほぼ 0 件になる。今の `unmanaged=57` は
ラボの既存リソースを拾っているノイズで、投影すると読めない。
`modified=1` だけが出る画面のほうが伝わる。

> ただし `tfdrift scan` が「state に無い実リソース全部」を unmanaged として
> 数えるなら、リージョン内の既存リソースは相変わらず拾う。**今夜実測して、
> unmanaged が減るかどうかを確認する。** 減らないなら、口頭では `modified` に
> だけ触れて `unmanaged` の数字は読まない。

## drift-sg.sh をこの SG に向ける

```bash
DEMO_SG=sg-xxxxxxxx DEMO_PORT=22 DEMO_CIDR=0.0.0.0/0 bash demo/drift-sg.sh
```

毎回打つのが面倒なら `demo/drift-sg.sh` の `DEMO_SG` 既定値を書き換える。

## 三者一致の確認（デモ前に毎回）

```bash
cd tf && terraform plan
```

**`No changes.` が出ることが、デモの出発点。** ここで差分が出ていると
「ドリフトを作った」のか「もともとズレていた」のかが観客に区別できない。

## 片付け（登壇後）

```bash
DEMO_SG=sg-xxxxxxxx DEMO_PORT=22 DEMO_CIDR=0.0.0.0/0 bash demo/undrift-sg.sh
cd tf && terraform destroy
```

**22番を開けたまま帰らない。** ラボでも開けっぱなしは開けっぱなし。

## 壇上で `terraform plan` も見せるなら

コンソールで 22 を開けたあとに `terraform plan` を打つと、差分が出る。
これは S3 のスライド（plan の出力）の実演になる:

- **plan でも見つかる** — 嘘をつかない
- ただし **誰かが打たないと出ない**（この瞬間、私が打ったから出た）
- そして **誰が開けたかは出ない**

尺が許せば強い beat になるが、`scan.sh` と役割が重なる。
**両方やるなら plan を先に、scan を後に。** 「plan と同じことを、exit code 付きで
CI に置ける形にしたのが scan」という順序になる。
