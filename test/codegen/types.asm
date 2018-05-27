	.data
a.0:		.word	0
b.1:		.asciiz "str"
c.2:		.word	0
d.3:		.asciiz ""
e.4:		.word	0
f.5:		.asciiz "Hello types"

	.text

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	sw	$3, a.0	# spilled a.0, freed $3
	li	$3, 0		# c.2 -> $3
	sw	$3, c.2	# spilled c.2, freed $3
	li	$3, 2		# e.4 -> $3
	# Store dirty variables back into memory
	sw	$3, e.4
	li	$2, 10
	syscall
	.end main
