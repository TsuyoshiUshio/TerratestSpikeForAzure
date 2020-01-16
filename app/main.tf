resource "helm_release" "nginx" {
    name = "my-nginx"
    chart = "nginx"
}