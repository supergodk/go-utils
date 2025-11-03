#!/bin/bash
# 先同步本地 tags
git fetch --tags --quiet

# 然后获取最新的 tag
latest_tag=$(git tag -l | grep -E '^v[0-9]' | sort -V | tail -1)

cat <<EOF
最新的 tag 是: $latest_tag
请输入要上传的 tag 名称:
EOF

read tag

if [ -z "$tag" ]; then
    echo "tag不能为空"
    exit 1
fi

git tag v$tag
git push origin v$tag