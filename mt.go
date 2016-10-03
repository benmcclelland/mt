/*
Package mt is a Go library for interacting with SCSI magnetic tape drives.
It wraps the mt executable and parses the output.  The mt executable
is readily available in most distros.  This library expects compatibility
with mt-st-1.1 in RedHat flavored distros
*/
package mt

import (
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Drive holds session information when interacting with a magnetic tape drive
type Drive struct {
	// Device is the device file in use for this Drive
	Device string
	// Command is the mt command used for the Drive
	Command string
	// Protects command exec
	mu sync.Mutex
}

// NewDrive returns a drive for a given device path
func NewDrive(device string) *Drive {
	return &Drive{Device: device, Command: "mt"}
}

// NewDriveCmd returns a Drive for a given device path and mt command
func NewDriveCmd(device, cmd string) *Drive {
	return &Drive{Device: device, Command: cmd}
}

// ForwardFiles forward space n files.
// The tape is positioned on the first block of the next file.
func (d *Drive) ForwardFiles(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "fsf", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "fsf")
}

// ForwardFileMarks forward space past n file marks,
// then backward space one file record.
// This leaves the tape positioned on the
// last block of the file that is n-1 files past the current file.
func (d *Drive) ForwardFileMarks(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "fsfm", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "fsfm")
}

// BackwardFiles backward space n files.
// The tape is positioned on the last block of the previous file.
func (d *Drive) BackwardFiles(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "bsf", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "bsf")
}

// BackwardFileMarks backward space past n file marks,
// then forward space one file record.
// This leaves the tape positioned on the first block of
// the file that is n-1 files before the current file.
func (d *Drive) BackwardFileMarks(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "bsfm", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "bsfm")
}

// PositionToFile the tape is positioned at the beginning of
// the nth file.
// Positioning is done by first rewinding the tape and then
// spacing forward over n filemarks.
func (d *Drive) PositionToFile(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "asf", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "asf")
}

// ForwardRecords forward space n records.
func (d *Drive) ForwardRecords(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "fsr", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "fsr")
}

// BackwardRecords backward space n records.
func (d *Drive) BackwardRecords(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "bsr", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "bsr")
}

// ForwardSetMarks (SCSI tapes) forward space n setmarks.
func (d *Drive) ForwardSetMarks(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "fss", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "fss")
}

// BackwardSetMarks (SCSI tapes) backward space n setmarks.
func (d *Drive) BackwardSetMarks(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "bss", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "bss")
}

// PositionEOD to end of valid data.
// Used on streamer tape drives to append data to the
// logical end of tape.
func (d *Drive) PositionEOD() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "eod")
	return errors.Wrap(err, "eod")
}

// Rewind the tape.
func (d *Drive) Rewind() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "rewind")
	return errors.Wrap(err, "rewind")
}

// Eject will rewind the tape and, if applicable,
// unload the tape.
func (d *Drive) Eject() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "eject")
	return errors.Wrap(err, "eject")
}

// Retension will wewind the tape, then wind it to the
// end of the reel, then rewind it again.
func (d *Drive) Retension() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "retension")
	return errors.Wrap(err, "retension")
}

// WriteEOFMarks write n EOF marks at current position.
func (d *Drive) WriteEOFMarks(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "weof", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "weof")
}

// WriteSetMarks (SCSI tapes) Write n setmarks at
// current position (only SCSI tape).
func (d *Drive) WriteSetMarks(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "wset", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "wset")
}

// Erase the tape.
func (d *Drive) Erase() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "erase")
	return errors.Wrap(err, "erase")
}

// Status will return status information about the tape unit.
func (d *Drive) Status() (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	result, err := mtCmd(d.Command, d.Device, "status")
	if err != nil {
		return "", errors.Wrap(err, "status")
	}
	return string(result[:]), nil
}

// SeekTape (SCSI tapes) seek to the nth block on the tape.
func (d *Drive) SeekTape(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "seek", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "seek")
}

// Tell (SCSI tapes) tell the current block on tape.
func (d *Drive) Tell() (string, error) {
	// TODO: return int64 instead of string
	d.mu.Lock()
	defer d.mu.Unlock()
	result, err := mtCmd(d.Command, d.Device, "tell")
	if err != nil {
		return "", errors.Wrap(err, "tell")
	}
	return string(result[:]), nil
}

// SetPartition (SCSI tapes) Switch to the nth partition. The
// default data partition of the tape is numbered zero. Switching
// partition  is available only if enabled for the device, the device
// supports multiple partitions, and the tape is formatted  with  multiple
// partitions.
func (d *Drive) SetPartition(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "setpartition", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "setpartition")
}

// SeekPartition (SCSI tapes) the tape position is set to nth block in the
// partition given by the argument.
func (d *Drive) SeekPartition(n, part int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "partseek",
		strconv.FormatInt(n, 10), strconv.FormatInt(part, 10))
	return errors.Wrap(err, "partseek")
}

// MakePartition (SCSI tapes) format the tape with one (n is zero) or two
// partitions (n gives the size of the second partition in megabytes).
// The tape drive must be able to format partitioned tapes with initiator
// specified partition size and partition support must be enabled for the drive.
func (d *Drive) MakePartition(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "mkpartition", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "mkpartition")
}

// Load (SCSI tapes) send the load command to the tape drive.
// The drives usually load the tape when a new cartridge is inserted.
func (d *Drive) Load() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "load")
	return errors.Wrap(err, "load")
}

// Lock (SCSI tapes) lock the tape drive door.
func (d *Drive) Lock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "lock")
	return errors.Wrap(err, "lock")
}

// Unlock (SCSI tapes) unlock the tape drive door.
func (d *Drive) Unlock() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "unlock")
	return errors.Wrap(err, "unlock")
}

// SetBlockSize (SCSI tapes) set the blocksize of the
// drive to n bytes per record.
func (d *Drive) SetBlockSize(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "setblk", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "setblk")
}

// SetDensity (SCSI tapes) set the tape density code to n.
// The proper codes to use with each drive should be looked
// up from the drive documentation.
func (d *Drive) SetDensity(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "setdensity", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "setdensity")
}

// SetDriveBuffer (SCSI tapes) set the tape drive buffer code to number.
// The proper value for unbuffered operation is zero and "normal" buffered
// operation one. The meanings of other values can be found in the drive
// documentation or, in the case of a SCSI-2 drive, from the SCSI-2 standard.
func (d *Drive) SetDriveBuffer(n int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "drvbuffer", strconv.Itoa(n))
	return errors.Wrap(err, "drvbuffer")
}

// SetCompression (SCSI tapes) the compression within the drive can be switched
// on or off using the MTCOMPRESSION ioctl. Note that this method is not supported
// by all drives implementing compression.
// arguments: true to enable, false to disable
func (d *Drive) SetCompression(enable bool) error {
	var state string
	if enable {
		state = "1"
	} else {
		state = "0"
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "compression", state)
	return errors.Wrap(err, "compression")
}

// StSetOptions (SCSI tapes) set the driver options bits for the device to the
// defined values. The bits can be set either by ORing the option bits from
// the file /usr/include/linux/mtio.h and passing in as a string, or by using
// the following keywords:
//   buffer-writes  buffered writes enabled
//   async-writes   asynchronous writes enabled
//   read-ahead     read-ahead for fixed block size
//   debug          debugging (if compiled into driver)
//   two-fms        write two filemarks when file closed
//   fast-eod       space directly to eod (and lose file number)
//   no-wait        don’t wait until rewind, etc. complete
//   auto-lock      automatically lock/unlock drive door
//   def-writes     the block size and density are for writes
//   can-bsr        drive can space backwards as well
//   no-blklimits   drive doesn’t support read block limits
//   can-partitions drive can handle partitioned tapes
//   scsi2logical   seek  and  tell  use  SCSI-2  logical block addresses
//                  instead of device dependent addresses
//   sili           Set the SILI bit is when reading  in  variable  block
//                  mode.  This  may speed up reading blocks shorter than
//                  the read byte count. Set this option only if you know
//                  that  the  drive  supports  SILI and the HBA reliably
//                  returns transfer residual byte counts. Requires  ker-
//                  nel version >= 2.6.26.
//   sysv           enable the System V semantics
func (d *Drive) StSetOptions(args ...string) error {
	optargs := append([]string{"stoptions"}, args...)
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, optargs...)
	return errors.Wrap(err, "stoptions")
}

// StClearOptions (SCSI tapes) clear selected driver option bits. The methods to
// specify the bits to clear are given above in description of StSetOptions.
func (d *Drive) StClearOptions(args ...string) error {
	optargs := append([]string{"stclearoptions"}, args...)
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, optargs...)
	return errors.Wrap(err, "stclearoptions")
}

// StShowOptions (SCSI tapes) print the currently enabled options for the device.
// Requires kernel version >= 2.6.26 and sysfs must be mounted at /sys.
func (d *Drive) StShowOptions() (string, error) {
	// TODO: return []string options
	d.mu.Lock()
	defer d.mu.Unlock()
	result, err := mtCmd(d.Command, d.Device, "stshowopt")
	if err != nil {
		return "", errors.Wrap(err, "stshowopt")
	}
	return string(result[:]), nil

}

// SetWriteThreashold (SCSI tapes) the write threshold for the tape device is
// set to n kilobytes. The value must be smaller than or equal to the driver
// buffer size.
func (d *Drive) SetWriteThreashold(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "stwrthreshold", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "stwrthreshold")
}

// SetDefaultBlockSize (SCSI tapes) set the default blocksize of the device to
// n bytes. The value -1 disables the default blocksize. The blocksize set by
// SetBlockSize overrides the default until a new tape is inserted.
func (d *Drive) SetDefaultBlockSize(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "defblksize", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "defblksize")
}

// SetDefaultDensity (SCSI tapes) set the default density code. The value -1
// disables the default density. The density set by SetDensity overrides the
// default until a new tape is inserted.
func (d *Drive) SetDefaultDensity(n int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "defdensity", strconv.FormatInt(n, 10))
	return errors.Wrap(err, "defdensity")
}

// SetDefaultDriveBuffer (SCSI tapes) set the default drive buffer code. The
// value -1 disables the default drive buffer code. The drive buffer code
// set by SetDriveBuffer overrides the default until a new tape is inserted.
func (d *Drive) SetDefaultDriveBuffer(n int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "defdrvbuffer", strconv.Itoa(n))
	return errors.Wrap(err, "defdrvbuffer")
}

// SetDefaultCompression (SCSI tapes) set the default compression state.
// The compression state set by SetCompression overrides the default until
// a new tape is inserted.
func (d *Drive) SetDefaultCompression(enable bool) error {
	var state string
	if enable {
		state = "1"
	} else {
		state = "0"
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "defcompression", state)
	return errors.Wrap(err, "defcompression")
}

// DisableDefaultCompression (SCSI tapes) disable the default compression state.
func (d *Drive) DisableDefaultCompression() error {
	state := "-1"
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "defcompression", state)
	return errors.Wrap(err, "defcompression")
}

// SetTimeout sets the normal timeout for the device.
// The value is given in seconds.
func (d *Drive) SetTimeout(n int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "sttimeout", strconv.Itoa(n))
	return errors.Wrap(err, "sttimeout")
}

// SetLongTimeout sets the long timeout for the device.
// The value is given in seconds.
func (d *Drive) SetLongTimeout(n int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "stlongtimeout", strconv.Itoa(n))
	return errors.Wrap(err, "stlongtimeout")
}

// SetClean set the cleaning request interpretation parameters.
func (d *Drive) SetClean() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := mtCmd(d.Command, d.Device, "stsetcln")
	return errors.Wrap(err, "stsetcln")
}

func mtCmd(mtcmd, dev string, args ...string) ([]byte, error) {
	cmdargs := append([]string{"-f", dev}, args...)
	cmd := exec.Command(mtcmd, cmdargs...)
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		err = errors.Wrap(err, "mt command setup stderr pipe")
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		err = errors.Wrap(err, "mt start command")
		return []byte{}, err
	}
	cmdout, err := ioutil.ReadAll(stdout)
	if err != nil {
		err = errors.Wrap(err, "mt read stdout output")
		return []byte{}, err
	}
	cmderr, err := ioutil.ReadAll(stderr)
	if err != nil {
		err = errors.Wrap(err, "mt read stderr output")
		return []byte{}, err
	}
	if err := cmd.Wait(); err != nil {
		err = errors.Wrap(err, "mt wait command")
		err = errors.Wrap(err, strings.TrimSuffix(string(cmderr), "\n"))
		return []byte{}, err
	}
	return cmdout, nil
}
