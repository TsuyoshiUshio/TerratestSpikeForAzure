resource "helm_release" "wordpress" {
    name = "my-wordpress"
    chart = "stable/wordpress"
}