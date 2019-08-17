load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "313f2c7a23fecc33023563f082f381a32b9b7254f727a7dd2d6380ccc6dfe09b",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/rules_go/releases/download/0.19.3/rules_go-0.19.3.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/0.19.3/rules_go-0.19.3.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "be9296bfd64882e3c08e3283c58fcb461fa6dd3c171764fcc4cf322f60615a9b",
    urls = [
        "https://storage.googleapis.com/bazel-mirror/github.com/bazelbuild/bazel-gazelle/releases/download/0.18.1/bazel-gazelle-0.18.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.18.1/bazel-gazelle-0.18.1.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

go_repository(
    name = "com_github_team_spectre_go_bigarray",
    importpath = "github.com/team-spectre/go-bigarray",
    sum = "h1:0XhUg6UMZlPvKOFKPKarKrRgMdDTJtXnL9r39nBzan4=",
    version = "v1.0.0",
)

go_repository(
    name = "com_github_team_spectre_go_bigbitvector",
    importpath = "github.com/team-spectre/go-bigbitvector",
    sum = "h1:EkeUZAOMA//b0t0Rn2LbeOy3NNSmwmNtg/ziEfTVl1I=",
    version = "v1.0.0",
)
