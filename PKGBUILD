pkgname=kbit-torrent
pkgver=0.0.0
pkgrel=1
pkgdesc="A minimal BitTorrent CLI client written in Go"
arch=('x86_64')
url="https://github.com/IdanKoblik/kbit-torrent"
license=('MIT')
depends=()
makedepends=('go')
source=("$pkgname-$pkgver.tar.gz::$url/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')

build() {
    cd "$pkgname-$pkgver"
    export CGO_ENABLED=0
    go build -trimpath \
        -ldflags="-s -w" \
        -o bin/kbit-torrent ./cmd
}

check() {
    cd "$pkgname-$pkgver"
    go test -short ./...
}

package() {
    cd "$pkgname-$pkgver"
    install -Dm755 bin/kbit-torrent "$pkgdir/usr/bin/kbit-torrent"
    install -Dm644 kbit-torrent.1 "$pkgdir/usr/share/man/man1/kbit-torrent.1"
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
}
