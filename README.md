# amblightd

A daemon which reads values from the (ACPI-based) Ambient Light Sensor, then sets the screen brightness accordingly using information from the built in response table (`\_SB.ALS._ALR`). It is not very configurable at the moment.

When the brightness is adjusted manually, the daemon takes this into account. In the future, the daemon will instead accept communication of manual brightness adjustments over a named pipe and adjust its internal state accordingly.

## Usage

`sudo amblightd`

It must be run as root.