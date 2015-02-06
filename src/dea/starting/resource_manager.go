package starting

import (
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/config"
	"syscall"
	"errors"
	"fmt"
)

//vm 物理机资源管理,主要关注 disk,memory
type ResourceManager struct {
	logger				*steno.Logger
	totalMemory		int
	totalDisk			int
	standbyMemory		int	//预留空虚内存
	standbyDisk		int //预留空虚磁盘
	diskPath			string //磁盘路径
}


//create resource manager
func NewResourceManager(config *config.Config) *ResourceManager {
	
	return &ResourceManager{
		logger:			steno.NewLogger("dea-docker"),
		totalMemory:		config.Dea.MemoryMb * config.Dea.MemoryFactor,
		totalDisk:			config.Dea.DiskMb * config.Dea.DiskFactor,
		standbyMemory:	config.Dea.StandbyMemory,
		standbyDisk:		config.Dea.StandbyDisk,
		diskPath:			config.Dea.DiskPath,
	}
}

//验证系统资源是否满足
func (r *ResourceManager) ValidateResource(memoryLimit int, diskLimit int) error {
	
	freeMemory := r.RemainingMemory()
	freeDisk   := r.RemainingDisk()
	
	if freeMemory < memoryLimit {
		return errors.New(fmt.Sprintf("Not enough memory resource available,free_memory: %s", freeMemory))
	}
	
	if freeDisk < diskLimit {
		return errors.New(fmt.Sprintf("Not enough disk resource available,free_disk: %s", freeMemory))
	}
	
	return nil
}


//获取剩余内存
func (r *ResourceManager) RemainingMemory () int {
	
	free := r.freeMemory()
	
	return (free - r.standbyMemory)
}

//获取剩余磁盘
func (r *ResourceManager) RemainingDisk () int {
	
	free := r.freeDisk()
	
	return (free - r.standbyDisk)
}

//return vm free memory
func (r *ResourceManager) freeMemory() int {

	sysInfo := new(syscall.Sysinfo_t)
    err := syscall.Sysinfo(sysInfo)
    if err != nil {
    	r.logger.Errorf("ResourceManager freeMemory fail,%s", err)
    	return 0
    }
    
	return (int(sysInfo.Freeram)/1024/1024)
}

//return vm total memory
func (r *ResourceManager) allMemory() int {

	sysInfo := new(syscall.Sysinfo_t)
    err := syscall.Sysinfo(sysInfo)
    if err != nil {
    	r.logger.Errorf("ResourceManager freeMemory fail,%s", err)
    	return 0
    }
    
	return (int(sysInfo.Totalram)/1024/1024)
}

//return vm free disk
func (r *ResourceManager) freeDisk() int {

	fs := syscall.Statfs_t{}
    err := syscall.Statfs(r.diskPath, &fs)
    if err != nil {
    	r.logger.Errorf("ResourceManager freeDisk fail,%s", err)
       return 0
    }
    free := fs.Bfree * uint64(fs.Bsize)
    
	return int(free/1024/1024)
}
