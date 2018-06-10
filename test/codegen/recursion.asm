	.data
n:		.word	0
newline.0:	.asciiz "\n"
t0:		.word	0
return.0:	.word	0
t1:		.word	0
PrintNatNums.0:	.word	0
x.1:		.word	0
t2:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime
PrintNatNums:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	lw	$3, PrintNatNums.0	# PrintNatNums.0 -> $3
	move	$5, $3		# n -> $5
	# Store dirty variables back into memory
	sw	$5, n
	bne	$5, 0, l0

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
	blt	$3, 1, l2

	li	$3, 0		# return.0 -> $3
	# Store dirty variables back into memory
	sw	$3, return.0

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end PrintNatNums
l2:
	li	$2, 1
	lw	$3, n		# n -> $3
	move	$4, $3
	syscall
	li	$2, 4
	la	$4, newline.0
	syscall
	sub	$5, $3, 1
	move	$6, $5		# PrintNatNums.0 -> $6
	sw	$3, n
	sw	$5, t1
	sw	$6, PrintNatNums.0
	jal	PrintNatNums
	lw	$5, t1
	lw	$6, PrintNatNums.0
	lw	$6, return.0	# return.0 -> $6
	move	$7, $6		# x.1 -> $7
	addi	$6, $7, 1
	move	$8, $6		# return.0 -> $8
	# Store dirty variables back into memory
	sw	$6, t2
	sw	$7, x.1
	sw	$8, return.0

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end PrintNatNums

	.globl main
	.ent main
main:
	li	$3, 5		# PrintNatNums.0 -> $3
	sw	$3, PrintNatNums.0
	jal	PrintNatNums
	lw	$5, t1
	lw	$6, PrintNatNums.0
	lw	$3, PrintNatNums.0
	li	$2, 10
	syscall
	.end main
