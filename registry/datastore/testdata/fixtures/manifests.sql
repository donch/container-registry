INSERT INTO "manifests"("id","schema_version","media_type","digest_hex","configuration_id","payload","created_at","marked_at","deleted_at")
VALUES
(1,2,E'application/vnd.docker.distribution.manifest.v2+json',decode('bd165db4bd480656a539e8e00db265377d162d6b98eebbfe5805d0fbd5144155','hex'),1,convert_to('{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1640,"digest":"sha256:ea8a54fd13889d3649d0a4e45735116474b8a650815a2cda4940f652158579b9"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":2802957,"digest":"sha256:c9b1b535fdd91a9855fb7f82348177e5f019329a58c53c47272962dd60f71fc9"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":108,"digest":"sha256:6b0937e234ce911b75630b744fb12836fe01bda5f7db203927edbb1390bc7e21"}]}', 'UTF8'),E'2020-03-02 17:50:26.461745',NULL,NULL),
(2,2,E'application/vnd.docker.distribution.manifest.v2+json',decode('56b4b2228127fd594c5ab2925409713bd015ae9aa27eef2e0ddd90bcb2b1533f','hex'),2,convert_to('{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1819,"digest":"sha256:9ead3a93fc9c9dd8f35221b1f22b155a513815b7b00425d6645b34d98e83b073"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":2802957,"digest":"sha256:c9b1b535fdd91a9855fb7f82348177e5f019329a58c53c47272962dd60f71fc9"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":108,"digest":"sha256:6b0937e234ce911b75630b744fb12836fe01bda5f7db203927edbb1390bc7e21"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":109,"digest":"sha256:f01256086224ded321e042e74135d72d5f108089a1cda03ab4820dfc442807c1"}]}', 'UTF8'),E'2020-03-02 17:50:26.461745',NULL,NULL),
(3,2,E'application/vnd.docker.distribution.manifest.v2+json',decode('bca3c0bf2ca0cde987ad9cab2dac986047a0ccff282f1b23df282ef05e3a10a6','hex'),3,convert_to('{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":6775,"digest":"sha256:33f3ef3322b28ecfc368872e621ab715a04865471c47ca7426f3e93846157780"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":27091819,"digest":"sha256:68ced04f60ab5c7a5f1d0b0b4e7572c5a4c8cce44866513d30d9df1a15277d6b"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":23882259,"digest":"sha256:c4039fd85dccc8e267c98447f8f1b27a402dbb4259d86586f4097acb5e6634af"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":203,"digest":"sha256:c16ce02d3d6132f7059bf7e9ff6205cbf43e86c538ef981c37598afd27d01efa"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":107,"digest":"sha256:a0696058fc76fe6f456289f5611efe5c3411814e686f59f28b2e2069ed9e7d28"}]}', 'UTF8'),E'2020-03-02 17:50:26.461745',NULL,NULL),
(4,1,E'application/vnd.docker.distribution.manifest.v1+json',decode('ea1650093606d9e76dfc78b986d57daea6108af2d5a9114a98d7198548bfdfc7','hex'),NULL,convert_to('{"schemaVersion":1,"name":"gitlab-org/gitlab-test/frontend","tag":"0.0.1","architecture":"amd64","fsLayers":[{"blobSum":"sha256:68ced04f60ab5c7a5f1d0b0b4e7572c5a4c8cce44866513d30d9df1a15277d6b"},{"blobSum":"sha256:c4039fd85dccc8e267c98447f8f1b27a402dbb4259d86586f4097acb5e6634af"},{"blobSum":"sha256:c16ce02d3d6132f7059bf7e9ff6205cbf43e86c538ef981c37598afd27d01efa"}],"history":[{"v1Compatibility":"{\"architecture\":\"amd64\",\"config\":{\"Hostname\":\"\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\"],\"ArgsEscaped\":true,\"Image\":\"sha256:74df73bb19fbfc7fb5ab9a8234b3d98ee2fb92df5b824496679802685205ab8c\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"container\":\"fb71ddde5f6411a82eb056a9190f0cc1c80d7f77a8509ee90a2054428edb0024\",\"container_config\":{\"Hostname\":\"fb71ddde5f64\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) \",\"CMD [\\\"/bin/sh\\\"]\"],\"ArgsEscaped\":true,\"Image\":\"sha256:74df73bb19fbfc7fb5ab9a8234b3d98ee2fb92df5b824496679802685205ab8c\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{}},\"created\":\"2020-03-23T21:19:34.196162891Z\",\"docker_version\":\"18.09.7\",\"id\":\"13787be01505ffa9179a780b616b953330baedfca1667797057aa3af67e8b39d\",\"os\":\"linux\",\"parent\":\"c6875a916c6940e6590b05b29f484059b82e19ca0eed100e2e805aebd98614b8\",\"throwaway\":true}"},{"v1Compatibility":"{\"id\":\"c6875a916c6940e6590b05b29f484059b82e19ca0eed100e2e805aebd98614b8\",\"created\":\"2020-03-23T21:19:34.027725872Z\",\"container_config\":{\"Cmd\":[\"/bin/sh -c #(nop) ADD file:0c4555f363c2672e350001f1293e689875a3760afe7b3f9146886afe67121cba in / \"]}}"}],"signatures":[{"header":{"jwk":{"crv":"P-256","kid":"SVNG:A2VR:TQJG:H626:HBKH:6WBU:GFBH:3YNI:425G:MDXK:ULXZ:CENN","kty":"EC","x":"daLesX_y73FSCFCaBuCR8offV_m7XEohHZJ9z-6WvOM","y":"pLEDUlQMDiEQqheWYVC55BPIB0m8BIhI-fxQBCH_wA0"},"alg":"ES256"},"signature":"mqA4qF-St1HTNsjHzhgnHBeN38ptKJOi4wSeH4xc_FCEPv0OchAUJC6v2gYTP4TwostmX-AB1_z3jo9G_ZuX5w","protected":"eyJmb3JtYXRMZW5ndGgiOjIxNTQsImZvcm1hdFRhaWwiOiJDbjAiLCJ0aW1lIjoiMjAyMC0wNC0xNVQwODoxMzowNVoifQ"}]}', 'UTF8'),E'2020-04-15 09:47:26.461413',NULL,NULL);
