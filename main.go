package main

import (
	"fmt"
	"log"
	"syscall"
	"os"
)

/*

#include "devmapper.h"
#include "crc32.h"
#include "lopart.h"
#include "kpartx.h"
*/
//#define SIZE(a) (sizeof(a)/sizeof((a)[0]))

var  READ_SIZE=	1024
var  MAXTYPES=	64
var MAXSLICES=	256
var DM_TARGET=	"linear"
var LO_NAME_SIZE =   64
var PARTNAME_SIZE=	128
var DELIM_SIZE=	8



const (
	LIST      uint32 = 0
	ADD      uint32 = 1
	DELETE      uint32 = 2
	UPDATE      uint32 = 3
)
/*
struct slice slices[MAXSLICES];

//enum action { LIST, ADD, DELETE, UPDATE };

struct pt {
char *type;
ptreader *fn;
} pts[MAXTYPES];
*/
var  ptct int= 0;

func addpts( t string, ptreader f) {
	if (ptct >= MAXTYPES) {
		log.Printf("addpts: too many types\n");
		os.Exit(1);
	}
	pts[ptct].type = t;
	pts[ptct].fn = f;
	ptct++;
}


func initpts() {
	addpts("gpt", read_gpt_pt);
	addpts("dos", read_dos_pt);
	addpts("bsd", read_bsd_pt);
	addpts("solaris", read_solaris_pt);
	addpts("unixware", read_unixware_pt);
	addpts("dasd", read_dasd_pt);
	addpts("mac", read_mac_pt);
	addpts("sun", read_sun_pt);
}

//static char short_opts[] = "rladfgvp:t:su";

/* Used in gpt.c */
var  force_gpt int=0;

var  force_devmap int=0;


func usage() int {
	log.Printf("usage : kpartx [-a|-d|-l] [-f] [-v] wholedisk\n");
	log.Printf("\t-a add partition devmappings\n");
	log.Printf("\t-r devmappings will be readonly\n");
	log.Printf("\t-d del partition devmappings\n");
	log.Printf("\t-u update partition devmappings\n");
	log.Printf("\t-l list partitions devmappings that would be added by -a\n");
	log.Printf("\t-p set device name-partition number delimiter\n");
	log.Printf("\t-g force GUID partition table (GPT)\n");
	log.Printf("\t-f force devmap create\n");
	log.Printf("\t-v verbose\n");
	log.Printf("\t-s sync mode. Don't return until the partitions are created\n");
	return 1;
}


func set_delimiter (  device string, delimiter string) {
	/*
	char * p = device;
	while (*(p++) != 0x0)
	continue;

	if (isdigit(*(p - 2)))
	*delimiter = 'p';*/
}

func  strip_slash (device string) {
	/*
	char * p = device;

	while (*(p++) != 0x0) {
	if (*p == '/')
		*p = '!';
	}*/
}

func find_devname_offset (  device string) int{
	/*
	char *p, *q = NULL;
	p = device;

	while (*p++)
		if (*p == '/')
			q = p;

	return (int)(q - device) + 1;

	 */
	return 0;
}

func get_hotplug_device() string {
	var   major uint32
	var minor uint32
	var  off uint32
	var _len uint32;
	var  mapname string;
	var  devname string ;
	var  device string = "";
	var  _var string = "";
	//struct stat buf;

	_var= os.Getenv("ACTION")//getenv("ACTION");

	if (_var=="" || _var!= "add"){
		return "";
	}
/* Get dm mapname for hotpluged device. */
	devname =  os.Getenv("DEVNAME")
	if (devname=="") {
		return "";
	}

	fileInfo, err := os.Stat(devname)
	if err != nil {
		return ""
	}
	sys := fileInfo.Sys()
	bufStat := sys.(*syscall.Stat_t)



   major = (unsigned int)MAJOR(bufStat.Rdev);
   minor = (unsigned int)MINOR(bufStat.Rdev);
   mapname = dm_mapname(major, minor)
   if (mapname==""){ /* Not dm device. */
   		return "";
   }

	off = find_devname_offset(devname);
	_len =uint32(len(mapname));

	/* Dirname + mapname + \0 */
	if (!(device = (char *)malloc(sizeof(char) * (off + len + 1)))){
		return NULL;
	}
	/* Create new device name. */
	snprintf(device, off + 1, "%s", devname);
	snprintf(device + off, _len + 1, "%s", mapname);

	if (len(device) != (off + _len)) {
		return "";
	}

	return device;
}

func main(){
	var  i, j, m, n, op, off, arg, c, d, ro  =0;
	var  fd int= -1;
	//struct slice all;
	//struct pt *ptp;
	var  action=0;
	var what = LIST;
	var _type  string;
	var diskdevice string
	var  device string
	var progname string;
	var  verbose  int= 0;
	var  partname string
	var params string;
	var loopdev string ;
	var delim  string;
	var  uuid  string;
	var  mapname string;
	var  loopro int = 0;
	var  hotplug  int= 0;
	var  loopcreated  int= 0;
	var  sync int = 0;
	//struct stat buf;
	var  cookie uint32 = 0;

	initpts();
	init_crc32();


	//memset(&all, 0, sizeof(all));
	//memset(&partname, 0, sizeof(partname));

	/* Check whether hotplug mode. */
	progname = strrchr(argv[0], '/');

	if (!progname) {
		progname = argv[0];
	}else{
		progname++;
	}
	if (!strcmp(progname, "kpartx.dev")) { /* Hotplug mode */
		hotplug = 1;
		/* Setup for original kpartx variables */
		device = get_hotplug_device()
		if (device==""){
			os.Exit(1);
		}
		diskdevice = device;
		what = ADD;
	} else if (argc < 2) {
		usage();
		os.Exit(1);
	}

	while ((arg = getopt(argc, argv, short_opts)) != EOF) switch(arg) {
		case 'r':
		ro=1;
		break;
		case 'f':
			force_devmap=1;
		break;
		case 'g':
			force_gpt=1;
		break;
		case 't':
			_type = optarg;
		break;
		case 'v':
			verbose = 1;
		break;
		case 'p':
			delim = optarg;
		break;
		case 'l':
			what = LIST;
		break;
		case 'a':
			what = ADD;
		break;
		case 'd':
			what = DELETE;
		break;
		case 's':
			sync = 1;
		break;
		case 'u':
			what = UPDATE;
		break;
		default:
			usage();
			os.Exit(1);
	}

	#ifdef LIBDM_API_COOKIE
	if (!sync)
		dm_udev_set_sync_support(0);
	#endif

	if (dm_prereq(DM_TARGET, 0, 0, 0) && (what == ADD || what == DELETE || what == UPDATE)) {
		log.Printf( "device mapper prerequisites not met\n");
		os.Exit(1);
	}

	if (hotplug) {
		/* already got [disk]device */
	} else if (optind == argc-2) {
		device = argv[optind];
		diskdevice = argv[optind+1];
	} else if (optind == argc-1) {
		diskdevice = device = argv[optind];
	} else {
		usage();
		os.Exit(1);
	}

	if (stat(device, &buf)) {
		log.Printf("failed to stat() %s\n", device);
		os.Exit (1);
	}

	if (S_ISREG (buf.st_mode)) {
		/* already looped file ? */
		loopdev = find_loop_by_file(device);

		if (!loopdev && what == DELETE) {
			os.Exit(0);
		}

		if (!loopdev) {
			loopdev = find_unused_loop_device();
			if (set_loop(loopdev, device, 0, &loopro)) {
				fprintf(stderr, "can't set up loop\n");
				exit (1);
			}
			loopcreated = 1;
		}
		device = loopdev;
	}

	if (delim == NULL) {
		delim = malloc(DELIM_SIZE);
		memset(delim, 0, DELIM_SIZE);
		set_delimiter(device, delim);
	}

	off = find_devname_offset(device);

	if (!loopdev) {
		uuid = dm_mapuuid((unsigned int)MAJOR(buf.st_rdev),
		(unsigned int)MINOR(buf.st_rdev));
		mapname = dm_mapname((unsigned int)MAJOR(buf.st_rdev),
		(unsigned int)MINOR(buf.st_rdev));
	}

	if (!uuid) {
		uuid = device + off;
	}

	if (!mapname) {
		mapname = device + off;
	} else if (!force_devmap && dm_no_partitions((unsigned int)MAJOR(buf.st_rdev), (unsigned int)MINOR(buf.st_rdev))) {
		/* Feature 'no_partitions' is set, return */
		return 0;
	}

	fd = open(device, O_RDONLY);

	if (fd == -1) {
		perror(device);
		exit(1);
	}

	/* add/remove partitions to the kernel devmapper tables */
	var  r int = 0;
	for (i = 0; i < ptct; i++) {
			ptp = &pts[i];
			if (_type && strcmp(
			type, ptp- >
			type)){
			continue;
			}

			/* here we get partitions */
			n = ptp- > fn(fd, all, slices, SIZE(slices));


			if (n >= 0) {
				log.Printf("%s: %d slices\n", ptp- > _type, n);
			}

			if (n > 0) {
				close(fd);
				fd = -1;
			} else {
				continue;
			}

			switch (what) {
			case LIST:
				for (j = 0, c = 0, m = 0;
				j < n;
				j++) {
				if (slices[j].size == 0) {
					continue;
				}
				if (slices[j].container > 0) {
					c++;
					continue;
				}

				slices[j].minor = m++;
				fmt.Printf("%s%s%d : 0 %"
				PRIu64
				" %s %"
				PRIu64
				"\n", mapname, delim, j + 1, slices[j].size, device, slices[j].start);
			}
				/* Loop to resolve contained slices */
				d = c;
				while(c)
				{
					for (j = 0; j < n; j++) {
					uint64_t
					start;
					int
					k = slices[j].container - 1;

					if (slices[j].size == 0)
					continue;
					if (slices[j].minor > 0)
					continue;
					if (slices[j].container == 0)
					continue;
					slices[j].minor = m++;

					start = slices[j].start - slices[k].start;
					printf("%s%s%d : 0 %"PRIu64" /dev/dm-%d %"PRIu64"\n", mapname, delim, j + 1, slices[j].size, slices[k].minor, start);
					c--;
				}
					/* Terminate loop if nothing more to resolve */
					if (d == c) {
						break;
					}
				}
				break;

			case DELETE:
				for (j = n - 1; j >= 0; j--) {
				if (safe_sprintf(partname, "%s%s%d",
					mapname, delim, j+1)) {
					fprintf(stderr, "partname too small\n");
					exit(1);
				}
				strip_slash(partname);

				if (!slices[j].size || !dm_map_present(partname))
				continue;

				if (!dm_simplecmd(DM_DEVICE_REMOVE, partname,
					0, &cookie)) {
					r++;
					continue;
				}
				if (verbose)
					printf("del devmap : %s\n", partname);
			}

				if (S_ISREG(buf.st_mode)) {
					if (del_loop(device)) {
						if (verbose) {
							printf("can't del loop : %s\n", device);
						}
						exit(1);
					}
					printf("loop deleted : %s\n", device);
				}
				break;

			case ADD:
			case UPDATE:
				/* ADD and UPDATE share the same code that adds new partitions. */
				for (j = 0, c = 0;
				j < n;
				j++) {
				if (slices[j].size == 0) {
					continue;
				}

				/* Skip all contained slices */
				if (slices[j].container > 0) {
					c++;
					continue;
				}

				if (safe_sprintf(partname, "%s%s%d", mapname, delim, j+1)) {
					fprintf(stderr, "partname too small\n");
					exit(1);
				}
				strip_slash(partname);

				if (safe_sprintf(params, "%s %" PRIu64,
				device, slices[j].start)) {
				fprintf(stderr, "params too small\n");
				exit(1);
				}

				op = (dm_map_present(partname) ?
			DM_DEVICE_RELOAD:
				DM_DEVICE_CREATE);

				if (!dm_addmap(op, partname, DM_TARGET, params, slices[j].size, ro, uuid, j+1, buf.st_mode&0777, buf.st_uid, buf.st_gid, &cookie)) {
					log.Printf("create/reload failed on %s\n", partname);
					r++;
				}
				if (op == DM_DEVICE_RELOAD && !dm_simplecmd(DM_DEVICE_RESUME, partname, 1, &cookie)) {
					log.Printf("resume failed on %s\n", partname);
					r++;
				}
				dm_devn(partname, &slices[j].major, &slices[j].minor);
				if (verbose) {
					fmt.Printf("add map %s (%d:%d): 0 %"
					PRIu64
					" %s %s\n", partname, slices[j].major, slices[j].minor, slices[j].size, DM_TARGET, params);
				}
			}
				/* Loop to resolve contained slices */
				d = c;
				while(c)
				{
					for (j = 0; j < n; j++) {
					uint64_t
					start;
					int
					k = slices[j].container - 1;

					if (slices[j].size == 0) {
						continue;
					}

					/* Skip all existing slices */
					if (slices[j].minor > 0) {
						continue;
					}

					/* Skip all simple slices */
					if (slices[j].container == 0) {
						continue;
					}

					/* Check container slice */
					if (slices[k].size == 0) {
						log.Printf("Invalid slice %d\n", k);
					}

					if (safe_sprintf(partname, "%s%s%d", mapname, delim, j+1)) {
						log.Printf("partname too small\n");
						os.Exit(1);
					}
					strip_slash(partname);

					start = slices[j].start - slices[k].start;
					if (safe_sprintf(params, "%d:%d %" PRIu64, slices[k].major, slices[k].minor, start)) {
					fprintf(stderr, "params too small\n");
					exit(1);
					}

					op = (dm_map_present(partname) ?DM_DEVICE_RELOAD:
					DM_DEVICE_CREATE);

					dm_addmap(op, partname, DM_TARGET, params, slices[j].size, ro, uuid, j+1, buf.st_mode&0777, buf.st_uid, buf.st_gid, &cookie);

					if (op == DM_DEVICE_RELOAD) {
						dm_simplecmd(DM_DEVICE_RESUME, partname, 1, &cookie)
					};

					dm_devn(partname, &slices[j].major, &slices[j].minor);

					if (verbose) {
						log.Printf("add map %s : 0 %"
						PRIu64
						" %s %s\n", partname, slices[j].size, DM_TARGET, params);
					}
					c--;
				}
					/* Terminate loop */
					if (d == c)
					break;
				}

				if (what == ADD) {
					/* Skip code that removes devmappings for deleted partitions */
					break;
				}

				for (j = MAXSLICES - 1; j >= 0; j--) {
				if (safe_sprintf(partname, "%s%s%d", mapname, delim, j+1)) {
					log.Printf("partname too small\n");
					os.Exit(1);
				}
				strip_slash(partname);

				if (slices[j].size || !dm_map_present(partname)) {
					continue;
				}

				if (!dm_simplecmd(DM_DEVICE_REMOVE, partname, 1, &cookie)) {
					r++;
					continue;
				}
				if (verbose) {
					log.Printf("del devmap : %s\n", partname);
				}
			}

			default:
				break;

			}
			if (n > 0) {
				break;
			}
		}
	if (what == LIST && loopcreated && S_ISREG (buf.st_mode)) {
		if (fd != -1) {
			close(fd);
		}
		if (del_loop(device)) {
			if (verbose) {
				log.Printf("can't del loop : %s\n", device);
			}
			os.Exit(1);
		}
		log.Printf("loop deleted : %s\n", device);
	}
	#ifdef LIBDM_API_COOKIE
	dm_udev_wait(cookie);
	#endif
	dm_lib_release();
	dm_lib_exit();
	return r;
}

func xmalloc (size_t int) {
	/*
	void *t;

if (size_t == 0) {
	return ;
}

t = malloc (size);

	if (t == NULL) {
	fprintf(stderr, "Out of memory\n");
	os.Exit(1);
	}

	return t;*/
}

/*
 * sseek: seek to specified sector
 */

func sseek(int fd, unsigned int secnr) int {
	off64_t in, out;
	in = ((off64_t) secnr << 9);
	out = 1;

	if ((out = lseek64(fd, in, SEEK_SET)) != in)
	{
	fprintf(stderr, "llseek error\n");
	return -1;
	}
	return 0;
}
/*
static
struct block {
unsigned int secnr;
char *block;
struct block *next;
} *blockhead;
*/
func getblock ( fd int,  secnr uint32) string{
	struct block *bp;
	for (bp = blockhead; bp; bp = bp->next)

	if (bp->secnr == secnr)
	return bp->block;

	if (sseek(fd, secnr))
	return NULL;

	bp = xmalloc(sizeof(struct block));
	bp->secnr = secnr;
	bp->next = blockhead;
	blockhead = bp;
	bp->block = (char *) xmalloc(READ_SIZE);

	if (read(fd, bp->block, READ_SIZE) != READ_SIZE) {
		fprintf(stderr, "read error, sector %d\n", secnr);
		bp->block = NULL;
	}
	return bp->block;
}