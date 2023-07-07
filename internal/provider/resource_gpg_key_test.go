package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccArgoCDGPGKeyResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `
resource "argocd_gpg_key" "this" {
	public_key = chomp(
<<-EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBGSJdlcBEACnza+KvWLyKWUHJPhgs//HRL0EEmA/EcFKioBlrgPNYf/O7hNg
KT3NDaNrD26pr+bOb4mfaqNNS9no8b9EP3C7Co3Wf2d4xpJ5/hlpIm3V652S5daZ
I7ylVT8QOrhaqEnHH2hEcOfDaqjrYfrx3qiI8v7DmV6jfGi1tDUUgfJwiOyZk4q1
jiPo5k4+XNp9mCtUAGyidLFcUqQ9XbHKgBwgAoxtIKNSbdPCGhsjgTHHhzswMH/Z
DhhtcraqrfOhoP9lI4/zyCS+B9OfUy7BS/1SqWKIgdsjFIR+zHIOI69lh77+ZAVE
MVYJBdFke5/g/tTPaQGuBqaIJ3d/Mi/ZlbTsoBcq5qam73uh7fcgBV5la6NeuNcR
tvKMVl4DlnkJS8LBtElLEeHEylTCdNltrUFwshDKDBtq6ilTKCK14R6g4lkn8VcE
9xx7Mhdh77tp66FRZ6ge1E8EUEFwEeFhp240KRyaA5U1/kAarn8083zZ7d4+QObp
L4KMqgrwLaxyPLgu0J/f946qLewV7XsbZRXE1jQa9Z7W5TEoJwjcC79DXe1wChc6
cBfCtluDsnklwvldpKTEZU0q/hKE6Zt7NjLUyExV+5guoHllxoVxx7sh+jtKm/J+
5gh+B3xOTDxRV2XYIx1TM6U1iLxAqchzFec8dfkuTbs/5f++PrddvZfiUQARAQAB
tD1BcmdvQ0QgVGVycmFmb3JtIFByb3ZpZGVyIDxmYWtldXNlckB1c2Vycy5ub3Jl
cGx5LmdpdGh1Yi5jb20+iQJOBBMBCgA4FiEEvK9bNlncXDhFAk6kmtkpVUAdOI0F
AmSJdlcCGwMFCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQmtkpVUAdOI2FdA//
YuFYsX6SUVgI4l68ZHE34jLTWU5R2ujB6luErcguAlLyDtrD3melva3V/ETc69/1
5o7Ayn3a7uz5lCEvUSLsCN+V2o3EjrA81pt8Zs+Z9WYeZE5F5DnKzq81PObdASB7
Po2X0qLqqKIhpQxc/E7m26xmePCf82H36gtvPiEVmVA5yduk1lLG3aZtNIRCa4VK
gmDjR8Se+OZeAw7JQCOeJB9/Y8oQ8nVkj1SWNIICaUwIXHtrj7r1z6XTDAEkGeBg
HXW8IEhZDE1Nq3vQtZvgwftEoPT/Ff+8DwvL1JUov2ObQDolallzKaiiVfGZhPJZ
4PMtEPEmSL9QWJAG5jiBVC3BdVZtXBNkC1HqTCXwZc/wzp5O9MmMXmCrUFr4FfHu
IZ560MNpp/SrtUrOahLmvuG0B+Ze96e2nm5ap5wkCDaQouOIqM7Lj+FGq64cu2B/
oSsl7joBZQUYXv8meNOQssm6jArRLG2oFoiEdRqzd2/RjvvJliLN9OCNvV43f38h
8Ep8RDi9RiHhSKvwrvDD9x/JRm6zQUetjrctmjdIYp8k129LrD0Qr9ULXfphZdrv
xga7/lyQLmukLu7Mxwp+ss2bY/wjT8mlT5P55kBpXXyYILhLsUESCHG6D8/Ov+vv
OoZS+BSfe/0vc1aTfDKxj5wAx27a6z5o25X27feEl3U=
=kqkH
-----END PGP PUBLIC KEY BLOCK-----
EOF
)
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "id", "9AD92955401D388D"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "fingerprint", "BCAF5B3659DC5C3845024EA49AD92955401D388D"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "owner", "ArgoCD Terraform Provider <fakeuser@users.noreply.github.com>"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "sub_type", "rsa4096"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "trust", "unknown"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "argocd_gpg_key.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update (i.e. recreate)
			{
				Config: `
resource "argocd_gpg_key" "this" {
	public_key = <<EOF
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBGSJpQYBEADIy7tAUiB7m5D159KphRN+E+5gc735v6wCqQz2IDpy1pvYpMeK
5N2MATUam6KAzRDTrfAPY4ztyFMTvzH3MuXZKgRFo/diUcYsT+lORRVgGELJQ4Jc
nC52vr5LmiLpTf4BQg8Er8dsCrLwPHUuUENbSigPPXtltIxYlkWRVutXcGKMuqNY
tYz8/UmT19dwsM5ZVyVbFfugymLP/ig2TT7s+oYIbTicoZZLtTXAP7oARK9NB23e
bOJUCsBJxyqGRor0AJ5pnkH18MYqWnrmEZrhkrsYdOS39+JWheadM1/w1v3CcDHa
prFiaWtuJovTTfa4NZi8982MV9IXAe1q/vTIjTWZDWNjyQhSASucNmHPwynczwTX
EnhqJn2G5STk+qnOcxEbcCcUx7eUEHPfHf1S+ubP9pjOCZhvwRiBNDxZAagp+Owj
9jC7m+AtCZ+qEoMKWMfm8KDmwt+1frfUSa4kMH6IX8OgLj1fcRSa0lGicuPf8N8z
1oEf2MvIg/Ey/vZGuj8nC/gafJuPOiwbLhSwXzPrXtdwvKjerGjJXFmxrBT2vVVc
UHUIPxktY5DAzwz3g2HKItNTuuK7fBwiT4a4XdkZweEU9ay3LeG10tnjbMianHhw
w0tQame9/fN4x/2UQGdwraOBWHlKtODcCdkGwh4OJGi/ONU0dRf4pjDfXwARAQAB
tD1BcmdvQ0QgVGVycmFmb3JtIFByb3ZpZGVyIDxmYWtldXNlckB1c2Vycy5ub3Jl
cGx5LmdpdGh1Yi5jb20+iQJOBBMBCgA4FiEEkIF5sfPAA3XyA8qv4a1Re5E3tjUF
AmSJpQYCGwMFCwkIBwIGFQoJCAsCBBYCAwECHgECF4AACgkQ4a1Re5E3tjWIdRAA
ll5uXni1Q8U+su5HkSt8sam46VJfXXDm1SGNZDXNDWEzUY+LXhtirpHMIAZH4sms
qph36okx+nrX7GowAP3NRYPgYibtP4Fbc1m19VmVxWNd7l7nr3k0apgsT+69tPRe
2ZwjFuLJeQJknGQnbPkPB22mvyG7zJv9JrVgo7nsYeIufoeCzFl2sx5c1uQS5NW3
D+eJP5s1ZmZj7fm/d8J1R/TX9Da527VUG1Q25bOuOxFoLpHMCNT0ABySLTOawCGU
4I3AURp1sHH78hT/X4gzwnADUPoVN6PfkUbkkVRoL88jpyI4AHIOJBKF/RwoC+AS
M3R7utv7JctVE3SkbWH0/4ihbV2mnYQWXSCgMLrJ4MT7pe4EvnZ4rHkZLEQxkGmR
MriiEvnqv4lhReygXw8bsciWm2KpqwxmJ2Vas7fMiIi//aAgzxLrC0cQCdc29KeV
pQS7vdV1OFghoL0y16OQ/ZIAdrmm11zKCl/iDxueOdxIuD2qKT+YKDVySOO4HPGD
3JRmebwT2CNmK09T8dp06LDovVUNvc7reBQ7aFsjhtvnzedoKhOuoIHyZVfMWklv
PtP70LwbQoVm073lQUsXiARC5UTA/RdqDgVrpgWtk6rs/uKHxkiOg20eMRGPaPV1
srrz5Qn8Ao4EGwXUs1Qg5yhCecqErcaSyaVNri7Jx2k=
=falI
-----END PGP PUBLIC KEY BLOCK-----
EOF
}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "id", "E1AD517B9137B635"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "fingerprint", "908179B1F3C00375F203CAAFE1AD517B9137B635"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "owner", "ArgoCD Terraform Provider <fakeuser@users.noreply.github.com>"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "sub_type", "rsa4096"),
					resource.TestCheckResourceAttr("argocd_gpg_key.this", "trust", "unknown"),
				),
			},
		},
	})
}

func TestAccArgoCDGPGKeyResource_Invalid_NotAGPGKey(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "argocd_gpg_key" "invalid" {
	public_key = "invalid"
}
				`,
				ExpectError: regexp.MustCompile("Invalid PGP Public Key"),
			},
		},
	})
}
