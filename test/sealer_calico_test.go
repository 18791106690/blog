package test

import (
	"blog/test/testhelper"
	"fmt"
	. "github.com/onsi/ginkgo"
	"strings"
	"time"

	"blog/test/suites/apply"
	"blog/test/testhelper/settings"
)

var _ = Describe("run calico ", func() {
	Context("start apply calico", func() {
		rawClusterFilePath := apply.GetRawClusterFilePath()
		rawCluster := apply.LoadClusterFileFromDisk(rawClusterFilePath)
		rawCluster.Spec.Image = settings.TestImageName
		rawCluster.Spec.Env = settings.CalicoEnv
		BeforeEach(func() {
			if rawCluster.Spec.Image != settings.TestImageName {
				rawCluster.Spec.Image = settings.TestImageName
				apply.MarshalClusterToFile(rawClusterFilePath, rawCluster)
			}
		})

		Context("check regular scenario that provider is bare metal, executes machine is master0", func() {
			var tempFile string
			BeforeEach(func() {
				tempFile = testhelper.CreateTempFile()
			})

			AfterEach(func() {
				testhelper.RemoveTempFile(tempFile)
			})
			It("init, clean up", func() {
				By("start to prepare infra")
				cluster := rawCluster.DeepCopy()
				cluster.Spec.Provider = settings.AliCloud
				cluster.Spec.Image = settings.TestImageName
				cluster = apply.CreateAliCloudInfraAndSave(cluster, tempFile)
				defer apply.CleanUpAliCloudInfra(cluster)
				sshClient := testhelper.NewSSHClientByCluster(cluster)
				testhelper.CheckFuncBeTrue(func() bool {
					err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
					return err == nil
				}, settings.MaxWaiteTime)

				By("start to init cluster")
				apply.GenerateClusterfile(tempFile)
				apply.SendAndApplyCluster(sshClient, tempFile)

				By("start to delete cluster")
				err := sshClient.SSH.CmdAsync(sshClient.RemoteHostIP, apply.SealerDeleteCmd(tempFile))
				testhelper.CheckErr(err)

				By("apply.SealerDelete()")
				time.Sleep(20 *time.Second)

				By("sealer run calico")
				masters := strings.Join(cluster.Spec.Masters.IPList, ",")
				nodes := strings.Join(cluster.Spec.Nodes.IPList, ",")
				apply.SendAndRunCluster(sshClient, tempFile, masters, nodes, cluster.Spec.SSH.Passwd)
				apply.CheckNodeNumWithSSH(sshClient, 2)
				fmt.Println("test finish")
				//
				//By("exec e2e test")
				////下载e2e镜像包
				//apply.GetE2eTest()
				////将kubernetes_e2e_images_v1.20.0.tar传输到孤岛环境，在每个k8s节点上执行docker load
				//err = sshClient.SSH.Copy(cluster.Spec.Masters.IPList[0], settings.LoadPath, settings.LoadPath)
				//if err != nil {
				//	return
				//}
				//err = sshClient.SSH.Copy(cluster.Spec.Nodes.IPList[0], settings.LoadPath, settings.LoadPath)
				//if err != nil {
				//	return
				//}
				//
				////进入到master节点并且执行load.sh
				//
				//_ = fmt.Sprintf("ssh root@%s", cluster.Spec.Masters.IPList[0])
				//_ = fmt.Sprintf("sh load.sh")
				//os.Exit(1)
				////进入到node节点
			})
		})
	})
})