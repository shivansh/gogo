	.data
a.0:		.word	0
binary.1:	.space	40
startSen.2:	.asciiz "Give input number less than 1024\n"
newline.3:	.asciiz "\n"
binaryLine.4:	.asciiz "The binary representation of the given number is \n"
i.5:		.word	0
t0:		.word	0
t1:		.word	0
t2:		.word	0
t3:		.word	0
t4:		.word	0
t5:		.word	0
j.6:		.word	0
t6:		.word	0
t7:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 0		# a.0 -> $3
	li	$2, 4
	la	$4, startSen.2
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	li	$2, 4
	la	$4, newline.3
	syscall
	li	$2, 4
	la	$4, binaryLine.4
	syscall
	sw	$3, a.0	# spilled a.0, freed $3
	li	$3, 0		# i.5 -> $3
	# Store dirty variables back into memory
	sw	$3, i.5

l2:
	lw	$3, a.0		# a.0 -> $3
	ble	$3, 0, l0

	li	$3, 1		# t0 -> $3
	# Store dirty variables back into memory
	sw	$3, t0
	j	l1

l0:
	li	$3, 0		# t0 -> $3
	# Store dirty variables back into memory
	sw	$3, t0

l1:
	lw	$3, t0		# t0 -> $3
	blt	$3, 1, l3

	la	$3, binary.1
	lw	$5, i.5		# i.5 -> $5
	sll	$s2, $5, 2	# iterator *= 4
	lw	$6, binary.1($s2)	# variable <- array
	sw	$6, t1		# spilled t1, freed $6
	lw	$6, a.0		# a.0 -> $6
	rem	$7, $6, 2
	move	$8, $7		# t1 -> $8
	sll $s2, $5, 2	# iterator *= 4
	sw	$8, binary.1($s2)	# variable -> array
	div	$9, $6, 2
	move	$6, $9		# a.0 -> $6
	addi	$10, $5, 1
	move	$5, $10		# i.5 -> $5
	# Store dirty variables back into memory
	sw	$5, i.5
	sw	$6, a.0
	sw	$7, t2
	sw	$8, t1
	sw	$9, t3
	sw	$10, t4
	j	l2

l3:
	lw	$3, i.5		# i.5 -> $3
	sub	$5, $3, 1
	move	$3, $5		# j.6 -> $3
	# Store dirty variables back into memory
	sw	$3, j.6
	sw	$5, t5

l6:
	lw	$3, j.6		# j.6 -> $3
	blt	$3, 0, l4

	li	$3, 1		# t6 -> $3
	# Store dirty variables back into memory
	sw	$3, t6
	j	l5

l4:
	li	$3, 0		# t6 -> $3
	# Store dirty variables back into memory
	sw	$3, t6

l5:
	lw	$3, t6		# t6 -> $3
	blt	$3, 1, l7

	la	$3, binary.1
	lw	$5, j.6		# j.6 -> $5
	sll	$s2, $5, 2	# iterator *= 4
	lw	$6, binary.1($s2)	# variable <- array
	li	$2, 1
	move	$4, $6
	syscall
	sub	$5, $5, 1
	# Store dirty variables back into memory
	sw	$5, j.6
	sw	$6, t7
	j	l6

l7:
	li	$2, 10
	syscall
	.end main
