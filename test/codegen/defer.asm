	.data
str.0:		.asciiz "First function call!\n"
a:		.word	0
c:		.word	0
str.1:		.asciiz "Second function call! Result: "
newline.2:	.asciiz "\n"
t0:		.word	0
sum.3:		.word	0
str.4:		.asciiz "Last function call!\n"
midFunc.0:	.word	0
midFunc.1:	.word	0
t3:		.word	0
t2:		.word	0
t1:		.word	0

	.text

	.globl firstFunc
	.ent firstFunc
firstFunc:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$2, 4
	la	$4, str.0
	syscall

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end firstFunc

	.globl midFunc
	.ent midFunc
midFunc:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	lw	$3, midFunc.0
	move	$5, $3		# a -> $5
	lw	$3, midFunc.1
	move	$6, $3		# c -> $6
	add	$3, $5, $6	# t0 -> $3
	move	$7, $3		# sum.3 -> $7
	li	$2, 4
	la	$4, str.1
	syscall
	li	$2, 1
	move	$4, $7
	syscall
	li	$2, 4
	la	$4, newline.2
	syscall
	# Store dirty variables back into memory
	sw	$3, t0
	sw	$5, a
	sw	$6, c
	sw	$7, sum.3

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end midFunc

	.globl lastFunc
	.ent lastFunc
lastFunc:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	li	$2, 4
	la	$4, str.4
	syscall

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end lastFunc

	.globl main
	.ent main
main:
	li	$3, 3		# midFunc.0 -> $3
	sw	$3, midFunc.0	# spilled midFunc.0, freed $3
	li	$3, 3		# midFunc.1 -> $3
	sw	$3, midFunc.1
	jal	firstFunc
	lw	$3, midFunc.1
	move	$3, $2
	sw	$3, t3
	jal	midFunc
	lw	$3, midFunc.1
	lw	$3, t3
	move	$3, $2
	sw	$3, t2
	jal	lastFunc
	lw	$3, midFunc.1
	lw	$3, t3
	lw	$3, t2
	move	$3, $2
	# Store dirty variables back into memory
	sw	$3, t1
	li	$2, 10
	syscall
	.end main
