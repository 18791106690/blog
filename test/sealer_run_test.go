package test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"strings"
	"time"

	"blog/test/suites/apply"
	"blog/test/testhelper"
	"blog/test/testhelper/settings"
)

var _ = Describe("sealer run", func() {
	fmt.Println("start to exec calico network cluster test")

	Context("run on bareMetal", func() {
		var tempFile string
		BeforeEach(func() {
			tempFile = testhelper.CreateTempFile()
		})

		AfterEach(func() {
			testhelper.RemoveTempFile(tempFile)
		})

		It("bareMetal run", func() {
			rawCluster := apply.LoadClusterFileFromDisk(apply.GetRawClusterFilePath())
			By("start to prepare infra")
			usedCluster := apply.CreateAliCloudInfraAndSave(rawCluster, tempFile)
			//defer to delete cluster
			defer func() {
				apply.CleanUpAliCloudInfra(usedCluster)
			}()
			sshClient := testhelper.NewSSHClientByCluster(usedCluster)
			testhelper.CheckFuncBeTrue(func() bool {
				err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
				return err == nil
			}, settings.MaxWaiteTime)

			By("start to init cluster", func() {
				masters := strings.Join(usedCluster.Spec.Masters.IPList, ",")
				nodes := strings.Join(usedCluster.Spec.Nodes.IPList, ",")
				apply.SendAndRunCluster(sshClient, tempFile, masters, nodes, usedCluster.Spec.SSH.Passwd)
				apply.SealerDelete()
				By("delete finish ==================")
				time.Sleep(20 *time.Second)
				apply.CheckNodeNumWithSSH(sshClient, 2)
			})
			Context("start apply hybridnet", func() {
				rawClusterFilePath := apply.GetRawClusterFilePath()
				rawCluster := apply.LoadClusterFileFromDisk(rawClusterFilePath)
				rawCluster.Spec.Image = settings.TestImageName
				rawCluster.Spec.Env = settings.HybridnetEnv
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
						//sshClient := testhelper.NewSSHClientByCluster(cluster)
						//testhelper.CheckFuncBeTrue(func() bool {
						//	err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
						//	return err == nil
						//}, settings.MaxWaiteTime)

						By("start to init cluster")
						apply.GenerateClusterfile(tempFile)
						apply.SendAndApplyCluster(sshClient, tempFile)
						apply.CheckNodeNumWithSSH(sshClient, 2)

						By("Wait for the cluster to be ready", func() {
							apply.WaitAllNodeRunningBySSH(sshClient.SSH,sshClient.RemoteHostIP)
						})
						By("start to delete cluster")
						err := sshClient.SSH.CmdAsync(sshClient.RemoteHostIP, apply.SealerDeleteCmd(tempFile))
						testhelper.CheckErr(err)
					})
				})
			})
			fmt.Println("calico network cluster test is ok")
		})
	})






	//fmt.Println("start to exec hybridnet network cluster test")
	//Context("run on bareMetal hybridnet", func() {
	//	var tempFile string
	//	BeforeEach(func() {
	//		tempFile = testhelper.CreateTempFile()
	//	})
	//
	//	AfterEach(func() {
	//		testhelper.RemoveTempFile(tempFile)
	//	})
	//
	//	It("bareMetal run", func() {
	//		rawCluster := apply.LoadClusterFileFromDisk(apply.GetRawClusterFilePath())
	//		By("start to prepare infra")
	//		usedCluster := apply.CreateAliCloudInfraAndSave(rawCluster, tempFile)
	//		//defer to delete cluster
	//		defer func() {
	//			apply.CleanUpAliCloudInfra(usedCluster)
	//		}()
	//		sshClient := testhelper.NewSSHClientByCluster(usedCluster)
	//		testhelper.CheckFuncBeTrue(func() bool {
	//			err := sshClient.SSH.Copy(sshClient.RemoteHostIP, settings.DefaultSealerBin, settings.DefaultSealerBin)
	//			return err == nil
	//		}, settings.MaxWaiteTime)
	//
	//		By("start to init cluster", func() {
	//			masters := strings.Join(usedCluster.Spec.Masters.IPList, ",")
	//			nodes := strings.Join(usedCluster.Spec.Nodes.IPList, ",")
	//			apply.SendAndRunHybirdnetCluster(sshClient, tempFile, masters, nodes, usedCluster.Spec.SSH.Passwd)
	//			apply.CheckNodeNumWithSSH(sshClient, 2)
	//		})
	//		By("Wait for the cluster to be ready", func() {
	//			apply.WaitAllNodeRunningBySSH(sshClient.SSH,sshClient.RemoteHostIP)
	//		})
	//		fmt.Println("calico network cluster test is ok")
	//	})
	//})

})
