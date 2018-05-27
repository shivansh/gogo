	.data
arr.0:		.space	8
t0:		.word	0
arr.1:		.space	8
t1:		.word	0
a.2:		.word	0
a.3:		.word	0

	.text

	.globl main
	.ent main
main:
	la	$3, arr.0
	lw	$5, 0($3)	# variable <- array
	li	$5, 4		# t0 -> $5
	sw	$5, 0($3)	# variable -> array
	sw	$5, t0		# spilled t0, freed $5
	la	$5, arr.1
	lw	$6, 0($5)	# variable <- array
	li	$6, 5		# t1 -> $6
	sw	$6, 0($5)	# variable -> array
	sw	$6, t1		# spilled t1, freed $6
	li	$6, 1		# a.2 -> $6
	sw	$6, a.2	# spilled a.2, freed $6
	li	$6, 4		# a.3 -> $6
	li	$6, 2		# a.3 -> $6
	sw	$6, a.3	# spilled a.3, freed $6
	li	$6, 4		# a.2 -> $6
	# Store dirty variables back into memory
	sw	$6, a.2
	li	$2, 10
	syscall
	.end main
